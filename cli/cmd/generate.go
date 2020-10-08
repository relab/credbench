package cmd

import (
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/relab/ct-eth-dapp/cli/datastore"
	"github.com/relab/ct-eth-dapp/cli/proto"
	"github.com/relab/ct-eth-dapp/cli/testconfig"
	"github.com/relab/ct-eth-dapp/src/ctree/aggregator"
	"github.com/relab/ct-eth-dapp/src/ctree/notary"

	pb "github.com/relab/ct-eth-dapp/cli/proto"
	keyutils "github.com/relab/ct-eth-dapp/src/accounts"
)

var (
	deployerAccount *pb.Account
	testConfig      testconfig.TestConfig
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate test case",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Reading test case configuration at:", testFile)
		var err error
		testConfig, err = testconfig.LoadConfig(testFile)
		if err != nil {
			log.Fatalln(err.Error())
		}
		err = setupTestCase()
		if err != nil {
			log.Fatalln(err.Error())
		}
	},
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run test case",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		testConfig, err = testconfig.LoadConfig(testFile)
		if err != nil {
			log.Fatalln(err.Error())
		}
		// call runTestCase
	},
}

func deployNotary(senderHexKey string) error {
	auth, err := keyutils.GetTxOpts(keyutils.HexToKey(senderHexKey), backend)
	if err != nil {
		return err
	}
	addr, tx, _, err := notary.DeployNotaryContract(auth, backend)
	if err != nil {
		return err
	}
	viper.Set("deployed_libs.notary", addr.Hex())
	viper.WriteConfig() //FIXME: this currently override the config
	fmt.Printf("Notary deployed at %s TXID: %x\n", addr.Hex(), tx.Hash())
	return nil
}

func deployAggregator(senderHexKey string) error {
	auth, err := keyutils.GetTxOpts(keyutils.HexToKey(senderHexKey), backend)
	if err != nil {
		return err
	}
	addr, tx, _, err := aggregator.DeployCredentialSum(auth, backend)
	if err != nil {
		return err
	}
	viper.Set("deployed_libs.aggregator", addr.Hex())
	viper.WriteConfig()
	fmt.Printf("Aggregator deployed at %s TXID: %x\n", addr.Hex(), tx.Hash())
	return nil
}

func setupTestCase() error {
	deployers, err := accountStore.GetByType(1, pb.Type_DEPLOYER)
	if err != nil || len(deployers) == 0 {
		return err
	}
	fmt.Println("Current deployers: ", deployers.ToHex())
	deployerAccount = deployers[0]

	deployNotary(deployerAccount.GetHexKey())

	deployAggregator(deployerAccount.GetHexKey())

	for f := 0; f < testConfig.Faculties; f++ {
		courses := make([]*proto.Course, testConfig.Courses)
		fAddr, err := createFaculty(testConfig.FacultyMembers)
		if err != nil {
			return err
		}
		for c := 0; c < testConfig.Courses; c++ {
			cAddr, err := createCourse(testConfig.Evaluators, testConfig.Students)
			if err != nil {
				return err
			}
			cs := datastore.NewCourseStore(db, cAddr)
			course, err := cs.GetCourse()
			courses = append(courses, course)

			err = enrollStudents(cAddr, testConfig.Students)
			if err != nil {
				return err
			}
		}
		// update faculty in db with courses addresses and diploma credentials
		fs := datastore.NewFacultyStore(db, fAddr)
		fs.SetCourses(courses)
	}
	return nil
}

func createCourse(e, s int) (common.Address, error) {
	evaluatorsAccounts, err := accountStore.GetAndSelect(e, pb.Type_EVALUATOR)
	if err != nil || len(evaluatorsAccounts) == 0 {
		return common.Address{}, err
	}
	evaluatorsAddresses := evaluatorsAccounts.ToETHAddress()
	key := keyutils.HexToKey(evaluatorsAccounts[0].GetHexKey())
	cAddr, _, err := DeployCourse(backend, key, evaluatorsAddresses, uint8(len(evaluatorsAccounts)))
	if err != nil {
		return common.Address{}, err
	}

	for _, ev := range evaluatorsAccounts {
		ev.ContractAddresses = append(ev.ContractAddresses, hexutil.Encode(cAddr.Bytes()))
	}

	c := &pb.Course{
		ContractAddress: hexutil.Encode(cAddr.Bytes()),
		Evaluators:      evaluatorsAccounts,
	}
	cs := datastore.NewCourseStore(db, cAddr)
	err = cs.PutCourse(c)
	if err != nil {
		return common.Address{}, err
	}

	return cAddr, nil
}

func enrollStudents(courseAddress common.Address, s int) error {
	studAccounts, err := accountStore.GetAndSelect(s, pb.Type_STUDENT)
	if err != nil {
		return err
	}

	for _, std := range studAccounts {
		std.ContractAddresses = append(std.ContractAddresses, hexutil.Encode(courseAddress.Bytes()))
	}

	for _, ss := range studAccounts {
		fmt.Println("STUD: ", ss)
	}

	cs := datastore.NewCourseStore(db, courseAddress)
	cs.SetStudents(studAccounts)

	// Get course contract
	contract, err := getCourse(courseAddress)
	if err != nil {
		return err
	}

	c, _ := cs.GetCourse()
	students := studAccounts.ToETHAddress()
	key := keyutils.HexToKey(c.Evaluators[0].GetHexKey())
	for _, student := range students {
		fmt.Println("ADDING STUDENT")
		fmt.Println("COURSE ADDRESS: ", contract.Address().Hex())
		_, err = addStudent(key, contract, student)
		// _, err := t.SendTX(senderHexKey, c.Address(), course.CourseContractABI, "addStudent", student)
		if err != nil {
			return err
		}
		fmt.Printf("Student %s successfully added\n", student.Hex())
	}

	b, err := contract.EnrolledStudents(&bind.CallOpts{Pending: false}, common.HexToAddress(studAccounts[0].HexAddress))
	fmt.Println(b, err)
	return nil
}

func createFaculty(a int) (common.Address, error) {
	admsAccounts, err := accountStore.GetAndSelect(a, pb.Type_ADM)
	if err != nil || len(admsAccounts) == 0 {
		return common.Address{}, err
	}
	admsAddresses := admsAccounts.ToETHAddress()
	key := keyutils.HexToKey(admsAccounts[0].GetHexKey())
	cAddr, _, err := DeployFaculty(backend, key, admsAddresses, uint8(len(admsAccounts)))
	if err != nil {
		return common.Address{}, err
	}

	for _, adms := range admsAccounts {
		adms.ContractAddresses = append(adms.ContractAddresses, hexutil.Encode(cAddr.Bytes()))
	}

	f := &pb.Faculty{
		ContractAddress: hexutil.Encode(cAddr.Bytes()),
		Adms:            admsAccounts,
	}

	fs := datastore.NewFacultyStore(db, cAddr)
	fs.AddFaculty(f)
	return cAddr, nil
}

// func runTestCase() {
// update course in db with published credentials
// 	//TODO: dispatch one go routine to each faculty and course
// 	for _, f := range testConfig.Faculties {
// 		// load faculty from DB
// 		for _, c := range f.Courses {
// 			// get info from DB
// 			// enrollStudents
// 			course, err := getCourse(cAddr)
// 			if err != nil {
// 				log.Fatalln(err.Error())
// 			}
// 			// for each evaluator
// 			// register credentials for all students
// 			// for each student
// 			// confirm credential
// 			// for each student, evaluator calls aggregate
// 			for _, s := c.NumberOfStudents {

// 			}
// 		}
// 		// for each course, adm perform
// 		// off-chain aggregation(query db hashes) and
// 		// calls register diploma with root
// 	}
// }
