package cmd

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math/big"
	"math/rand"
	"path/filepath"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/relab/ct-eth-dapp/bench/datastore"
	"github.com/relab/ct-eth-dapp/bench/genesis"
	"github.com/relab/ct-eth-dapp/bench/testconfig"
	"github.com/relab/ct-eth-dapp/bench/transactor"
	"github.com/relab/ct-eth-dapp/src/deployer"
	"github.com/relab/ct-eth-dapp/src/faculty"
	"github.com/relab/ct-eth-dapp/src/fileutils"
	"github.com/relab/ct-eth-dapp/src/schemes"

	pb "github.com/relab/ct-eth-dapp/bench/proto"
	course "github.com/relab/ct-eth-dapp/src/course"
)

var testConfig testconfig.TestConfig

// Generate the test case by deploying the certification tree.
// It deploy faculty and course contracts, and assign evaluators/owners.
var generateTestCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate test case",
	Run: func(cmd *cobra.Command, args []string) {
		log.Infoln("Reading test case configuration at:", testFile)
		var err error
		testConfig, err = testconfig.LoadConfig(testFile)
		if err != nil {
			log.Fatal(err)
		}

		err = setupTestCase()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func setupTestCase() error {
	opts, err := accountStore.GetTxOpts(defaultSender.Bytes(), backend)
	if err != nil {
		return err
	}
	err = deployNotary(opts, backend)
	if err != nil {
		return err
	}
	opts, err = accountStore.GetTxOpts(defaultSender.Bytes(), backend)
	if err != nil {
		return err
	}
	err = deployAggregator(opts, backend)
	if err != nil {
		return err
	}
	log.Debugln("Successfully deployed all libraries")

	eg := new(errgroup.Group)
	for f := 0; f < testConfig.Faculties; f++ {
		eg.Go(func() error {
			return setupFaculties()
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

func setupFaculties() error {
	admsAccounts, err := accountStore.GetAndSelect(testConfig.FacultyMembers, pb.Type_ADM)
	if err != nil || len(admsAccounts) == 0 {
		log.Debug(err, "...reusing existing accounts")
		admsAccounts, err = selectAccounts(testConfig.AccountDistribution, testConfig.FacultyMembers, pb.Type_ADM)
		if err != nil {
			return err
		}
	}

	fAddr, err := createFaculty(admsAccounts)
	if err != nil {
		return err
	}

	contract, err := getFacultyContract(fAddr)
	if err != nil {
		return err
	}

	for s := 0; s < testConfig.Semesters; s++ {
		semesterID := sha256.Sum256([]byte(fmt.Sprintf("semester-%d", s)))
		courses, err := createSemester(fAddr, semesterID)
		if err != nil {
			return err
		}

		// Add courses to faculty contract
		for _, c := range courses {
			opts, err := accountStore.GetTxOpts(admsAccounts[0].Address, backend)
			if err != nil {
				return err
			}

			_, err = contract.AddNode(opts, c)
			if err != nil {
				return err
			}
		}

		opts, err := accountStore.GetTxOpts(admsAccounts[0].Address, backend)
		if err != nil {
			return err
		}

		tx, err := contract.RegisterSemester(opts, semesterID, courses)
		if err != nil {
			return err
		}
		log.Debugf("semester %x successfully registered at tx: %s", semesterID, tx.Hash().Hex())
	}
	return nil
}

func createFaculty(admsAccounts datastore.Accounts) (common.Address, error) {
	opts, err := accountStore.GetTxOpts(defaultSender.Bytes(), backend)
	// opts, err := accountStore.GetTxOpts(admsAccounts[0].Address, backend)
	if err != nil {
		return common.Address{}, err
	}
	fAddr, _, err := DeployFaculty(opts, backend, admsAccounts.ToETHAddress(), uint8(len(admsAccounts)))
	if err != nil {
		return common.Address{}, err
	}

	for _, adms := range admsAccounts {
		adms.Contracts = append(adms.Contracts, fAddr.Bytes())
	}
	err = accountStore.PutAccount(admsAccounts...)
	if err != nil {
		return common.Address{}, err
	}

	f := &pb.Faculty{
		Address:   fAddr.Bytes(),
		Adms:      admsAccounts.ToBytes(),
		CreatedOn: timestamppb.Now(),
	}

	fs := datastore.NewFacultyStore(db, fAddr)
	err = fs.PutFaculty(f)
	if err != nil {
		return common.Address{}, err
	}
	return fAddr, nil
}

func createSemester(fAddr common.Address, semester [32]byte) ([]common.Address, error) {
	courseCh := make(chan common.Address)
	quit := make(chan struct{}, 1)
	var courses []common.Address

	defer func() {
		close(courseCh)
		close(quit)
	}()

	g := new(errgroup.Group)
	for s := 0; s < testConfig.Courses; s++ {
		g.Go(func() error {
			return registerCourse(courseCh)
		})
	}
	go func() {
		for course := range courseCh {
			courses = append(courses, course)
			if len(courses) == testConfig.Courses {
				quit <- struct{}{}
			}
		}
	}()
	if err := g.Wait(); err != nil {
		return []common.Address{}, err
	}

	// update faculty in db with courses addresses and diploma credentials
	fs := datastore.NewFacultyStore(db, fAddr)
	err := fs.AddSemester(semester)
	if err != nil {
		return []common.Address{}, err
	}

	<-quit
	return courses, nil
}

func registerCourse(courseCh chan common.Address) error {
	evaluatorsAccounts, err := accountStore.GetAndSelect(testConfig.Evaluators, pb.Type_EVALUATOR)
	if err != nil || len(evaluatorsAccounts) == 0 {
		log.Debug(err, "...reusing existing accounts")
		evaluatorsAccounts, err = selectAccounts(testConfig.AccountDistribution, testConfig.Evaluators, pb.Type_EVALUATOR)
		if err != nil {
			return err
		}
	}

	cAddr, err := createCourse(evaluatorsAccounts)
	if err != nil {
		return err
	}
	cs := datastore.NewCourseStore(db, cAddr)
	studAccounts, err := accountStore.GetAndSelect(testConfig.Students, pb.Type_STUDENT)
	if err != nil {
		log.Debug(err, "...reusing existing accounts")
		studAccounts, err = selectAccounts(testConfig.AccountDistribution, testConfig.Students, pb.Type_STUDENT)
		if err != nil {
			return err
		}
	}

	err = registerStudents(cs, cAddr, studAccounts)
	if err != nil {
		return err
	}

	courseCh <- cAddr
	return nil
}

func createCourse(evaluatorsAccounts datastore.Accounts) (common.Address, error) {
	evaluatorsAddresses := evaluatorsAccounts.ToETHAddress()
	opts, err := accountStore.GetTxOpts(defaultSender.Bytes(), backend)
	// opts, err := accountStore.GetTxOpts(evaluatorsAccounts[0].Address, backend)
	if err != nil {
		return common.Address{}, err
	}

	cAddr, _, err := DeployCourse(opts, backend, evaluatorsAddresses, uint8(len(evaluatorsAccounts)))
	if err != nil {
		return common.Address{}, err
	}

	// Append contract address for all evaluators
	for _, ev := range evaluatorsAccounts {
		ev.Contracts = append(ev.Contracts, cAddr.Bytes())
	}
	err = accountStore.PutAccount(evaluatorsAccounts...)
	if err != nil {
		return common.Address{}, err
	}

	// Append contract address for all evaluators
	c := &pb.Course{
		Address:    cAddr.Bytes(),
		Evaluators: evaluatorsAccounts.ToBytes(),
		CreatedOn:  timestamppb.Now(),
	}
	cs := datastore.NewCourseStore(db, cAddr)
	err = cs.PutCourse(c)
	if err != nil {
		return common.Address{}, err
	}

	return cAddr, nil
}

func registerStudents(cs *datastore.CourseStore, courseAddress common.Address, studAccounts datastore.Accounts) error {
	for _, std := range studAccounts {
		std.Contracts = append(std.Contracts, courseAddress.Bytes())
	}
	err := accountStore.PutAccount(studAccounts...)
	if err != nil {
		return err
	}

	err = cs.SetStudents(studAccounts)
	if err != nil {
		return err
	}
	return nil
}

// returns a random subset of keys of size n
func selectRandom(n int, keys [][]byte) [][]byte {
	exists := make(map[int]struct{})
	var chosen [][]byte
	for i := 0; i < n; {
		pos := rand.Intn(len(keys))
		if _, ok := exists[pos]; !ok {
			chosen = append(chosen, keys[pos])
			exists[pos] = struct{}{}
			i++
		}
	}
	return chosen
}

// returns a subset of keys of size n in a sequential order, starting from
// the given index
func selectSequentialFrom(n int, index int, keys [][]byte) ([][]byte, error) {
	chosen := make([][]byte, n)
	if index > len(keys)-1 || index < 0 {
		return nil, fmt.Errorf("invalid index")
	} else if index+n > len(keys)-1 {
		copy(chosen, keys[index:])
		copy(chosen[len(keys)-index:], keys[0:n-(len(keys)-index)])
	} else {
		copy(chosen, keys[index:index+n])
	}
	return chosen, nil
}

func selectKeys(method string, n int, keys [][]byte) ([][]byte, error) {
	var err error
	if n > len(keys) {
		return nil, fmt.Errorf("insufficient available keys")
	}

	switch method {
	case "random":
		keys = selectRandom(n, keys)
	case "sequential": // starting from random index
		keys, err = selectSequentialFrom(n, rand.Intn(len(keys)), keys)
		if err != nil {
			return nil, err
		}
	default:
		keys = keys[:n]
	}
	return keys, nil
}

func selectAccounts(method string, n int, selectType pb.Type) (datastore.Accounts, error) {
	keys, err := accountStore.GetAllKeys(selectType)
	if err != nil {
		return nil, err
	}
	if len(keys) < 1 {
		return nil, fmt.Errorf("insufficient number of accounts")
	}

	keys, err = selectKeys(method, n, keys)
	if err != nil {
		return nil, err
	}

	accounts, err := accountStore.GetAccounts(keys...)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

// Run the test case by enrolling students and producing a credential tree
// for them for the specified period.
var runTestCmd = &cobra.Command{
	Use:   "run",
	Short: "Run test case",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		testConfig, err = testconfig.LoadConfig(testFile)
		if err != nil {
			log.Fatal(err)
		}
		err = runTestCase()
		if err != nil {
			log.Fatal(err)
		}
	},
}

// Running test case
func runTestCase() error {
	done := make(chan struct{}, testConfig.Faculties)
	keys, err := db.Keys("faculties")
	if err != nil {
		return err
	}

	for _, key := range keys {
		go run_faculty(key, done)
	}

	quit := 0
	for range done {
		quit++
		if quit == testConfig.Faculties {
			close(done)
		}
	}
	return nil
}

func run_faculty(key []byte, done chan struct{}) {
	ldir := filepath.Join(logdir, fmt.Sprintf("/faculty_%x", key))
	err := fileutils.CreateDir(ldir)
	if err != nil {
		log.Fatal(err)
	}

	fs := datastore.NewFacultyStore(db, common.BytesToAddress(key))
	f, err := fs.GetFaculty()
	if err != nil {
		log.Fatal(err)
	}

	// Semesters are necessarily sequential
	wg := &sync.WaitGroup{}
	wg.Add(len(f.Semesters))
	for i, s := range f.Semesters {
		runner := transactor.NewTransactor(backend)

		go func() {
			defer wg.Done()
			gasLimit := big.NewInt(0)
			if _, ok := gasLimit.SetString(genesis.GasLimit, 10); !ok {
				log.Fatal("Error setting the gas limit")
			}
			gasPrice, _ := new(big.Int).SetString("20000000000", 10)
			logger := transactor.NewTXLogger(gasLimit, gasPrice)
			logFilename := filepath.Join(ldir, fmt.Sprintf("log_%d_%x.log", i, s))
			err := logger.SaveMetric(logFilename, runner.Metrics)
			if err != nil {
				log.Fatal(err)
			}
		}()

		students, err := runSemester(runner, s, f)
		if err != nil {
			log.Fatal(err)
		}
		if len(students) == 0 {
			log.Fatal("no students registered")
		}
		err = fs.AddStudent(students...)
		if err != nil {
			log.Error(err)
		}
		runner.Close()
	}
	wg.Wait()
	done <- struct{}{}
}

func runSemester(runner *transactor.Transactor, semester []byte, faculty *pb.Faculty) ([]common.Address, error) {
	facultyContract, err := getFacultyContract(common.BytesToAddress(faculty.Address))
	if err != nil {
		return []common.Address{}, err
	}

	courses, err := facultyContract.GetCoursesBySemester(&bind.CallOpts{Pending: false}, semester)
	if err != nil {
		return []common.Address{}, err
	}

	g := new(errgroup.Group)
	for _, cAddr := range courses {
		cs := datastore.NewCourseStore(db, cAddr)
		c, err := cs.GetCourse()
		if err != nil {
			return []common.Address{}, err
		}
		g.Go(func() error {
			return runCourse(runner, c)
		})
	}
	if err := g.Wait(); err != nil {
		return []common.Address{}, err
	}

	adms, err := accountStore.GetAccounts(faculty.Adms...)
	if err != nil {
		return []common.Address{}, err
	}

	students, err := issueSemesterCredential(runner, facultyContract, adms, courses)
	if err != nil {
		return []common.Address{}, err
	}

	aggregateSemesters(runner, facultyContract, faculty.Adms[0], students)

	return students, nil
}

func issueSemesterCredential(runner *transactor.Transactor, contract *faculty.Faculty, adms datastore.Accounts, courses []common.Address) ([]common.Address, error) {
	// Collect all students courses
	studentPerCourse := make(map[common.Address][]common.Address)
	for _, c := range courses {
		cc, err := getCourseContract(c)
		if err != nil {
			return []common.Address{}, err
		}
		students, err := cc.GetStudents(nil)
		if err != nil {
			return []common.Address{}, err
		}
		for _, s := range students {
			studentPerCourse[s] = append(studentPerCourse[s], c)
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(len(studentPerCourse))
	students := make([]common.Address, 0, len(studentPerCourse))
	for s, w := range studentPerCourse {
		students = append(students, s)

		go func(student common.Address, witnesses []common.Address) {
			defer wg.Done()

			var digest [32]byte
			for i, adm := range adms {
				if i == 0 {
					digest = schemes.GenerateRandomDigest(student.Bytes(), 32)
				}

				opts, err := accountStore.GetTxOpts(adm.Address, backend)
				if err != nil {
					log.Error(err)
				}

				tx, err := registerSemesterCredential(runner, opts, contract, student, digest, witnesses)
				if err != nil {
					log.Error(err)
				}

				err = deployer.WaitTxConfirmation(context.Background(), backend, tx, 0)
				if err != nil {
					log.Error(err)
				}
			}
			opts, err := accountStore.GetTxOpts(student.Bytes(), backend)
			if err != nil {
				log.Error(err)
			}

			tx, err := approveSemesterCredential(runner, opts, contract, digest)
			if err != nil {
				log.Error(err)
			}

			err = deployer.WaitTxConfirmation(context.TODO(), backend, tx, 0)
			if err != nil {
				log.Error(err)
			}
		}(s, w)
	}
	wg.Wait()
	return students, nil
}

// FIXME: DRY (courses/faculties)
func aggregateSemesters(runner *transactor.Transactor, contract *faculty.Faculty, adm []byte, students []common.Address) {
	wgs := sync.WaitGroup{}
	wgs.Add(len(students))
	for _, student := range students {
		student := student
		go func() {
			defer wgs.Done()

			digests, err := contract.GetDigests(nil, student)
			if err != nil {
				log.Error(err)
			}

			opts, err := accountStore.GetTxOpts(adm, backend)
			if err != nil {
				log.Error(err)
			}

			_, err = aggregateSemesterCredentials(runner, opts, contract, student, digests)
			if err != nil {
				log.Error(err)
			}
		}()
	}
	wgs.Wait()
}

func runCourse(runner *transactor.Transactor, cc *pb.Course) error {
	if len(cc.Evaluators) == 0 {
		return fmt.Errorf("No evaluators found")
	}

	students, err := accountStore.GetAccounts(cc.Students...)
	if err != nil {
		return err
	}
	if len(students) == 0 {
		return fmt.Errorf("No student found")
	}

	contract, err := getCourseContract(common.BytesToAddress(cc.Address))
	if err != nil {
		return fmt.Errorf("No course contract found")
	}

	evaluators, err := accountStore.GetAccounts(cc.Evaluators...)
	if err != nil {
		return err
	}

	// Enroll all students to the course contract
	err = enrollStudents(runner, contract, evaluators[0], students.ToETHAddress())
	if err != nil {
		return err
	}
	// Issue all exams for all students
	issueExams(runner, contract, evaluators, students)
	// Aggregate all exams for all students
	aggregateExams(runner, contract, evaluators[0], students.ToETHAddress())
	return nil
}

func enrollStudent(runner *transactor.Transactor, contract *course.Course, evaluator *pb.Account, student common.Address) (*types.Transaction, error) {
	opts, err := accountStore.GetTxOpts(evaluator.Address, backend)
	if err != nil {
		return nil, err
	}
	return addStudent(runner, opts, contract, student)
}

func enrollStudents(runner *transactor.Transactor, contract *course.Course, evaluator *pb.Account, students []common.Address) error {
	for i, student := range students {
		tx, err := enrollStudent(runner, contract, evaluator, student)
		if err != nil {
			return err
		}
		// Wait successful enrollment of the last student
		if i == len(students)-1 {
			err = deployer.WaitTxConfirmation(context.TODO(), backend, tx, 0)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// FIXME: ignored for now
// func generateExamCredential(registrar common.Address, student common.Address, course common.Address) [32]byte {
// 	courseEntity := &schemes.Entity{
// 		Id:   course.Hex(),
// 		Name: "Course Test Contract",
// 	}
// 	ag := schemes.NewFakeAssignmentGrade(registrar.Hex(), student.Hex())
// 	credential := schemes.NewFakeAssignmentGradeCredential(registrar.Hex(), courseEntity, ag)
// 	return schemes.Hash(credential)
// }

func issueExams(runner *transactor.Transactor, contract *course.Course, evaluators datastore.Accounts, students datastore.Accounts) {
	wgs := sync.WaitGroup{}
	wgs.Add(len(students))
	for _, s := range students {
		go func(student *pb.Account) {
			defer wgs.Done()
			studentAddress := common.BytesToAddress(student.Address)
			for e := 0; e < testConfig.Exams; e++ {
				var digest [32]byte
				for i, evaluator := range evaluators {
					// eAddress := common.BytesToAddress(evaluator.Address)
					if i == 0 {
						digest = schemes.GenerateRandomDigest(studentAddress.Bytes(), 32)
						// digest = generateExamCredential(eAddress, studentAddress, contract.Address())
					}

					opts, err := accountStore.GetTxOpts(evaluator.Address, backend)
					if err != nil {
						log.Fatal(err)
					}

					tx, err := registerCourseCredential(runner, opts, contract, studentAddress, digest)
					if err != nil {
						log.Error(err)
					}

					err = deployer.WaitTxConfirmation(context.Background(), backend, tx, 0)
					if err != nil {
						log.Error(err)
					}
				}

				opts, err := accountStore.GetTxOpts(student.Address, backend)
				if err != nil {
					log.Error(err)
				}

				tx, err := approveCourseCredential(runner, opts, contract, digest)
				if err != nil {
					log.Error(err)
				}

				err = deployer.WaitTxConfirmation(context.TODO(), backend, tx, 0)
				if err != nil {
					log.Error(err)
				}
			}
		}(s)
	}
	wgs.Wait()
}

func aggregateExams(runner *transactor.Transactor, contract *course.Course, evaluator *pb.Account, students []common.Address) {
	wgs := sync.WaitGroup{}
	wgs.Add(len(students))
	for _, student := range students {
		student := student
		go func() {
			defer wgs.Done()

			digests, err := contract.GetDigests(nil, student)
			if err != nil {
				log.Error(err)
			}

			opts, err := accountStore.GetTxOpts(evaluator.Address, backend)
			if err != nil {
				log.Error(err)
			}

			_, err = aggregateCourseCredentials(runner, opts, contract, student, digests)
			if err != nil {
				log.Error(err)
			}
		}()
	}
	wgs.Wait()
}

func newTestCmd() *cobra.Command {
	testCmd := &cobra.Command{
		Use:   "test",
		Short: "Manage tests",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			rootCmd.PersistentPreRun(cmd, args)
			err := loadDefaultAccount()
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	testCmd.AddCommand(
		generateTestCmd,
		runTestCmd,
		newGenAccountsCmd(),
	)
	return testCmd
}