package cmd

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"

	"github.com/relab/ct-eth-dapp/src/accounts"
	course "github.com/relab/ct-eth-dapp/src/course"
	pb "github.com/relab/ct-eth-dapp/src/schemes"
)

var addStudentCmd = &cobra.Command{
	Use:   "addStudent",
	Short: "Add a student to the course contract",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := getCourse(common.HexToAddress(args[0]))
		if err != nil {
			log.Fatalln(err.Error())
		}
		studentAddress := common.HexToAddress(args[1])
		key := wallet.PrivateKey()
		tx, err := addStudent(key, c, studentAddress)
		if err != nil {
			log.Fatalln(err.Error())
		}
		fmt.Printf("Transaction ID: %x\n", tx.Hash())
	},
}

func addStudent(senderKey *ecdsa.PrivateKey, c *course.Course, studentAddress common.Address) (*types.Transaction, error) {
	opts, _ := accounts.GetTxOpts(senderKey, backend)

	tx, err := c.AddStudent(opts, studentAddress)
	if err != nil {
		return nil, fmt.Errorf("Failed to add student: %v", err)
	}

	if ok, _ := c.IsEnrolled(&bind.CallOpts{Pending: false}, studentAddress); ok {
		fmt.Printf("student %s successfully enrolled!\n", studentAddress.Hex())
	}
	return tx, nil
}

var rmStudentCmd = &cobra.Command{
	Use:   "rmStudent",
	Short: "Remove a student to the course contract",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := getCourse(common.HexToAddress(args[0]))
		if err != nil {
			log.Fatalln(err.Error())
		}
		studentAddress := common.HexToAddress(args[1])
		key := wallet.PrivateKey()
		tx, err := rmStudent(key, c, studentAddress)
		if err != nil {
			log.Fatalln(err.Error())
		}
		fmt.Printf("Transaction ID: %x\n", tx.Hash())
	},
}

func rmStudent(senderKey *ecdsa.PrivateKey, c *course.Course, studentAddress common.Address) (*types.Transaction, error) {
	opts, _ := accounts.GetTxOpts(senderKey, backend)

	tx, err := c.RemoveStudent(opts, studentAddress)
	if err != nil {
		return nil, fmt.Errorf("Failed to remove student: %v", err)
	}

	if ok, _ := c.IsEnrolled(&bind.CallOpts{Pending: false}, studentAddress); !ok {
		fmt.Printf("student %s successfully removed!\n", studentAddress.Hex())
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
		c, err := getCourse(common.HexToAddress(args[0]))
		if err != nil {
			log.Fatalln(err.Error())
		}
		studentAddress := common.HexToAddress(args[1])
		a := &pb.AssignmentGradeCredential{}
		pb.ParseJSON(args[2], a)
		digest := pb.Hash(a)
		fmt.Println("ISSUING CREDENTIAL")
		key := wallet.PrivateKey()
		tx, err := registerCredential(key, c, studentAddress, digest)
		if err != nil {
			log.Fatalln(err.Error())
		}
		fmt.Printf("Transaction ID: %x\n", tx.Hash())
	},
}

func registerCredential(senderKey *ecdsa.PrivateKey, c *course.Course, studentAddress common.Address, digest [32]byte) (*types.Transaction, error) {
	opts, err := accounts.GetTxOpts(senderKey, backend)
	if err != nil {
		return nil, err
	}

	tx, err := c.RegisterCredential0(opts, studentAddress, digest, []common.Address{})
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func getCourse(courseAddress common.Address) (*course.Course, error) {
	c, err := course.NewCourse(courseAddress, backend)
	if err != nil {
		return nil, fmt.Errorf("Failed to get course: %v", err)
	}
	return c, nil
}

func newCourseCmd() *cobra.Command {
	courseCmd := &cobra.Command{
		Use:   "course",
		Short: "Course contract",
	}
	courseCmd.AddCommand(
		addStudentCmd,
		rmStudentCmd,
		issueCourseCredentialCmd,
	)
	return courseCmd
}
