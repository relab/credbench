package cmd

import (
	"errors"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"

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
			log.Fatalln(err.Error())
		}
		studentAddress := common.HexToAddress(args[1])

		opts, err := wallet.GetTxOpts(backend)
		if err != nil {
			log.Fatalln(err.Error())
		}

		tx, err := addStudent(opts, c, studentAddress)
		if err != nil {
			log.Fatalln(err.Error())
		}
		fmt.Printf("Transaction ID: %x\n", tx.Hash())
	},
}

func addStudent(opts *bind.TransactOpts, c *course.Course, studentAddress common.Address) (tx *types.Transaction, err error) {
	tx, err = executor.SendTX(opts, c.Address(), course.CourseContractABI, "addStudent", studentAddress)
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
			log.Fatalln(err.Error())
		}
		studentAddress := common.HexToAddress(args[1])

		opts, err := wallet.GetTxOpts(backend)
		if err != nil {
			log.Fatalln(err.Error())
		}

		tx, err := rmStudent(opts, c, studentAddress)
		if err != nil {
			log.Fatalln(err.Error())
		}
		fmt.Printf("Transaction ID: %x\n", tx.Hash())
	},
}

func rmStudent(opts *bind.TransactOpts, c *course.Course, studentAddress common.Address) (*types.Transaction, error) {
	tx, err := executor.SendTX(opts, c.Address(), course.CourseContractABI, "removeStudent", studentAddress)
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
			log.Fatalln(err.Error())
		}

		// Student is using the default wallet
		opts, err := wallet.GetTxOpts(backend)
		if err != nil {
			log.Fatalln(err.Error())
		}

		tx, err := renounce(opts, c, opts.From)
		if err != nil {
			log.Fatalln(err.Error())
		}
		fmt.Printf("Transaction ID: %x\n", tx.Hash())
	},
}

func renounce(opts *bind.TransactOpts, c *course.Course, studentAddress common.Address) (*types.Transaction, error) {
	tx, err := executor.SendTX(opts, c.Address(), course.CourseContractABI, "renounceCourse", studentAddress)
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
			log.Fatalln(err.Error())
		}
		studentAddress := common.HexToAddress(args[1])
		a := &pb.AssignmentGradeCredential{}
		pb.ParseJSON(args[2], a)
		digest := pb.Hash(a)

		opts, err := wallet.GetTxOpts(backend)
		if err != nil {
			log.Fatalln(err.Error())
		}

		tx, err := registerCredential(opts, c, studentAddress, digest)
		if err != nil {
			log.Fatalln(err.Error())
		}
		fmt.Printf("Transaction ID: %x\n", tx.Hash())
	},
}

func registerCredential(opts *bind.TransactOpts, c *course.Course, studentAddress common.Address, digest [32]byte) (*types.Transaction, error) {
	tx, err := executor.SendTX(opts, c.Address(), course.CourseContractABI, "registerCredential", studentAddress, digest, []common.Address{})
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
			log.Fatalln(err.Error())
		}
		students, err := c.GetStudents(&bind.CallOpts{Pending: false})
		if err != nil {
			log.Fatalln(err.Error())
		}
		fmt.Printf("Registered students:\n")
		for _, s := range students {
			fmt.Printf("%s\n", s.Hex())
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
			log.Fatalln(err.Error())
		}
		student := common.HexToAddress(args[1])
		ok, err := c.IsEnrolled(&bind.CallOpts{Pending: false}, student)
		if err != nil {
			log.Fatalln(err.Error())
		}
		if ok {
			fmt.Printf("Student %s is enrolled\n", student.Hex())
		} else {
			fmt.Printf("Student %s not found in course %s\n", student.Hex(), c.Address().Hex())
		}
	},
}

func newCourseCmd() *cobra.Command {
	courseCmd := &cobra.Command{
		Use:   "course",
		Short: "Manage course",
	}
	courseCmd.AddCommand(
		addStudentCmd,
		rmStudentCmd,
		renounceCourseCmd,
		getStudentsCmd,
		isEnrolledCmd,
		issueCourseCredentialCmd,
	)
	return courseCmd
}
