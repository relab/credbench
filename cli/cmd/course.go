package cmd

import (
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"

	"github.com/relab/ct-eth-dapp/cli/datastore"
	"github.com/relab/ct-eth-dapp/cli/transactor"
	course "github.com/relab/ct-eth-dapp/src/course"

	pb "github.com/relab/ct-eth-dapp/src/schemes"
)

var addStudentCmd = &cobra.Command{
	Use:   "addStudent",
	Short: "Add a student to the course contract",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := getCourseContract(common.HexToAddress(args[0]))
		if err != nil {
			log.Fatal(err)
		}
		studentAddress := common.HexToAddress(args[1])

		opts, err := accountStore.GetTxOpts(defaultSender.Bytes(), backend)
		if err != nil {
			log.Fatal(err)
		}

		tx, err := addStudent(executor, opts, c, studentAddress)
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("Transaction ID: %x\n", tx.Hash())
	},
}

func addStudent(e *transactor.Transactor, opts *bind.TransactOpts, c *course.Course, studentAddress common.Address) (tx *types.Transaction, err error) {
	tx, err = e.SendTX("course", opts, c.Address(), course.CourseContractABI, "addStudent", studentAddress)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

var rmStudentCmd = &cobra.Command{
	Use:   "rmStudent",
	Short: "Remove a student to the course contract",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := getCourseContract(common.HexToAddress(args[0]))
		if err != nil {
			log.Fatal(err)
		}
		studentAddress := common.HexToAddress(args[1])

		opts, err := accountStore.GetTxOpts(defaultSender.Bytes(), backend)
		if err != nil {
			log.Fatal(err)
		}

		tx, err := rmStudent(executor, opts, c, studentAddress)
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("Transaction ID: %x\n", tx.Hash())
	},
}

func rmStudent(e *transactor.Transactor, opts *bind.TransactOpts, c *course.Course, studentAddress common.Address) (*types.Transaction, error) {
	tx, err := e.SendTX("course", opts, c.Address(), course.CourseContractABI, "removeStudent", studentAddress)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

var renounceCourseCmd = &cobra.Command{
	Use:   "renounce",
	Short: "Renounce from a course",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := getCourseContract(common.HexToAddress(args[0]))
		if err != nil {
			log.Fatal(err)
		}

		// Note: student is considered to be using the default wallet
		opts, err := accountStore.GetTxOpts(defaultSender.Bytes(), backend)
		if err != nil {
			log.Fatal(err)
		}

		tx, err := renounce(executor, opts, c, opts.From)
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("Transaction ID: %x\n", tx.Hash())
	},
}

func renounce(e *transactor.Transactor, opts *bind.TransactOpts, c *course.Course, studentAddress common.Address) (*types.Transaction, error) {
	tx, err := e.SendTX("course", opts, c.Address(), course.CourseContractABI, "renounceCourse", studentAddress)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

var issueCourseCredentialCmd = &cobra.Command{
	Use:   "issue",
	Short: "Issue a new credential",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 3 {
			return errors.New("Missing arguments. Please specify: course_address student_address path_to_json_credential")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		c, err := getCourseContract(common.HexToAddress(args[0]))
		if err != nil {
			log.Fatal(err)
		}
		studentAddress := common.HexToAddress(args[1])
		a := &pb.AssignmentGradeCredential{}
		pb.ParseJSON(args[2], a)
		digest := pb.Hash(a)

		opts, err := accountStore.GetTxOpts(defaultSender.Bytes(), backend)
		if err != nil {
			log.Fatal(err)
		}

		tx, err := registerCourseCredential(executor, opts, c, studentAddress, digest)
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("Transaction ID: %x\n", tx.Hash())
	},
}

func registerCourseCredential(e *transactor.Transactor, opts *bind.TransactOpts, c *course.Course, studentAddress common.Address, digest [32]byte) (*types.Transaction, error) {
	tx, err := e.SendTX("course", opts, c.Address(), course.CourseContractABI, "registerCredential", studentAddress, digest, []common.Address{})
	if err != nil {
		return nil, err
	}
	return tx, nil
}

var approveCourseCredentialCmd = &cobra.Command{
	Use:   "approve",
	Short: "Approve a credential using its hash",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("Missing arguments. Please specify: course_address digest_hash")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		c, err := getCourseContract(common.HexToAddress(args[0]))
		if err != nil {
			log.Fatal(err)
		}
		digest := common.HexToHash(args[1]) // 0x?
		// FIXME: Validate inputs

		opts, err := accountStore.GetTxOpts(defaultSender.Bytes(), backend)
		if err != nil {
			log.Fatal(err)
		}

		tx, err := approveCourseCredential(executor, opts, c, digest)
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("Transaction ID: %x\n", tx.Hash())
	},
}

func approveCourseCredential(e *transactor.Transactor, opts *bind.TransactOpts, c *course.Course, digest [32]byte) (*types.Transaction, error) {
	tx, err := e.SendTX("course", opts, c.Address(), course.CourseContractABI, "approveCredential", digest)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func aggregateCourseCredentials(e *transactor.Transactor, opts *bind.TransactOpts, c *course.Course, student common.Address, digests [][32]byte) (*types.Transaction, error) {
	tx, err := e.SendTX("course", opts, c.Address(), course.CourseContractABI, "aggregateCredentials", student, digests)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func getCourseContract(courseAddress common.Address) (*course.Course, error) {
	c, err := course.NewCourse(courseAddress, backend)
	if err != nil {
		return nil, fmt.Errorf("Failed to get course: %v", err)
	}
	return c, nil
}

var getStudentsCmd = &cobra.Command{
	Use:   "getStudents",
	Short: "Return the list of enrolled students",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := getCourseContract(common.HexToAddress(args[0]))
		if err != nil {
			log.Fatal(err)
		}
		students, err := c.GetStudents(&bind.CallOpts{Pending: false})
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("Registered students:\n")
		for _, s := range students {
			log.Infof("%s\n", s.Hex())
		}
	},
}

var isEnrolledCmd = &cobra.Command{
	Use:   "isEnrolled",
	Short: "Check whether a student is enrolled in a course",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := getCourseContract(common.HexToAddress(args[0]))
		if err != nil {
			log.Fatal(err)
		}
		student := common.HexToAddress(args[1])
		ok, err := c.IsEnrolled(&bind.CallOpts{Pending: false}, student)
		if err != nil {
			log.Fatal(err)
		}
		if ok {
			log.Infof("Student %s is enrolled\n", student.Hex())
		} else {
			log.Infof("Student %s not found in course %s\n", student.Hex(), c.Address().Hex())
		}
	},
}

var getCourseCmd = &cobra.Command{
	Use:   "get",
	Short: "Shows the course details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		address := common.HexToAddress(args[0])
		cs := datastore.NewCourseStore(db, address)
		course, err := cs.GetCourse()
		if err != nil {
			log.Error(err)
		}
		fmt.Printf("Course Info:\n")
		fmt.Printf("\tAddress: %s\n", address.Hex())
		fmt.Printf("\tCreated on: %s\n", course.CreatedOn.AsTime())
		fmt.Printf("\tEvaluators:\n")
		for _, e := range course.Evaluators {
			fmt.Printf("\t  %s\n", common.BytesToAddress(e).Hex())
		}
		fmt.Printf("\tStudents:\n")
		for _, s := range course.Students {
			fmt.Printf("\t  %s\n", common.BytesToAddress(s).Hex())
		}
	},
}

func newCourseCmd() *cobra.Command {
	courseCmd := &cobra.Command{
		Use:   "course",
		Short: "Manage course",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			rootCmd.PersistentPreRun(cmd, args)
			err := loadDefaultAccount()
			if err != nil {
				log.Fatal(err)
			}
		},
	}
	courseCmd.AddCommand(
		addStudentCmd,
		rmStudentCmd,
		renounceCourseCmd,
		getStudentsCmd,
		isEnrolledCmd,
		issueCourseCredentialCmd,
		approveCourseCredentialCmd,
		getCourseCmd,
	)
	return courseCmd
}
