package cmd

import (
	"errors"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"

	"github.com/relab/ct-eth-dapp/src/core/accounts"
	"github.com/relab/ct-eth-dapp/src/core/course"
	contract "github.com/relab/ct-eth-dapp/src/core/course/contract"
	pb "github.com/relab/ct-eth-dapp/src/schemes"
)

var courseCmd = &cobra.Command{
	Use:   "course",
	Short: "Course contract",
}

func deployCourseCmd() *cobra.Command {
	var owners []string
	var quorum int64

	c := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy course contract",
		Run: func(cmd *cobra.Command, args []string) {
			tx, err := deployCourse(owners, quorum)
			if err != nil {
				log.Fatalln(err.Error())
			}
			fmt.Printf("Transaction ID: %x\n", tx.Hash())
		},
	}

	c.Flags().StringSliceVar(&owners, "owners", []string{}, "Owners addresses (comma separated)")
	c.Flags().Int64Var(&quorum, "quorum", int64(len(owners)), "Minimum number of signatures required to issue course credentials")

	c.MarkFlagRequired("owners")
	c.MarkFlagRequired("quorum")

	return c
}

func deployCourse(ownersList []string, quorum int64) (tx *types.Transaction, err error) {
	var owners []common.Address
	for _, addr := range ownersList {
		owners = append(owners, common.HexToAddress(addr))
	}

	backend, _ := clientConn.Backend()
	opts, _ := accounts.GetTxOpts(wallet.PrivateKey(), backend)

	var cAddr common.Address
	cAddr, tx, _, err = contract.DeployCourse(opts, backend, owners, big.NewInt(quorum))
	courseAddress := cAddr.Hex()
	if err != nil || courseAddress == "0x0000000000000000000000000000000000000000" {
		return nil, fmt.Errorf("failed to deploy the contract: %v", err)
	}
	fmt.Printf("Contract %v successfully deployed\n", courseAddress)
	return tx, nil
}

var addStudentCmd = &cobra.Command{
	Use:   "addStudent",
	Short: "Add a student to the course contract",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		course, err := getCourse(common.HexToAddress(args[0]))
		if err != nil {
			log.Fatalln(err.Error())
		}
		studentAddress := common.HexToAddress(args[1])
		tx, err := addStudent(course, studentAddress)
		if err != nil {
			log.Fatalln(err.Error())
		}
		fmt.Printf("Transaction ID: %x\n", tx.Hash())
	},
}

func addStudent(course *course.Course, studentAddress common.Address) (*types.Transaction, error) {
	backend, _ := clientConn.Backend()
	opts, _ := accounts.GetTxOpts(wallet.PrivateKey(), backend)

	tx, err := course.AddStudent(opts, studentAddress)
	if err != nil {
		return nil, fmt.Errorf("Failed to add student: %v", err)
	}

	if ok, _ := course.IsEnrolled(&bind.CallOpts{Pending: false}, studentAddress); ok {
		fmt.Printf("student %s successfully enrolled!\n", studentAddress.Hex())
	}
	return tx, nil
}

var rmStudentCmd = &cobra.Command{
	Use:   "rmStudent",
	Short: "Remove a student to the course contract",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		course, err := getCourse(common.HexToAddress(args[0]))
		if err != nil {
			log.Fatalln(err.Error())
		}
		studentAddress := common.HexToAddress(args[1])
		tx, err := rmStudent(course, studentAddress)
		if err != nil {
			log.Fatalln(err.Error())
		}
		fmt.Printf("Transaction ID: %x\n", tx.Hash())
	},
}

func rmStudent(course *course.Course, studentAddress common.Address) (*types.Transaction, error) {
	backend, _ := clientConn.Backend()
	opts, _ := accounts.GetTxOpts(wallet.PrivateKey(), backend)

	tx, err := course.RemoveStudent(opts, studentAddress)
	if err != nil {
		return nil, fmt.Errorf("Failed to remove student: %v", err)
	}

	if ok, _ := course.IsEnrolled(&bind.CallOpts{Pending: false}, studentAddress); !ok {
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
		course, err := getCourse(common.HexToAddress(args[0]))
		if err != nil {
			log.Fatalln(err.Error())
		}
		studentAddress := common.HexToAddress(args[1])
		a := &pb.AssignmentGradeCredential{}
		pb.ParseJSON(args[2], a)
		digest := pb.Hash(a)
		tx, err := registerCredential(course, studentAddress, digest)
		if err != nil {
			log.Fatalln(err.Error())
		}
		fmt.Printf("Transaction ID: %x\n", tx.Hash())
	},
}

func registerCredential(course *course.Course, studentAddress common.Address, digest [32]byte) (*types.Transaction, error) {
	backend, _ := clientConn.Backend()
	opts, _ := accounts.GetTxOpts(wallet.PrivateKey(), backend)

	tx, err := course.RegisterCredential(opts, studentAddress, digest)
	if err != nil {
		return nil, fmt.Errorf("Failed to register credential: %v", err)
	}
	return tx, nil
}

func getCourse(courseAddress common.Address) (*course.Course, error) {
	backend, _ := clientConn.Backend()
	course, err := course.NewCourse(courseAddress, backend)
	if err != nil {
		return nil, fmt.Errorf("Failed to get course: %v", err)
	}
	return course, nil
}

func init() {
	courseCmd.AddCommand(deployCourseCmd())
	courseCmd.AddCommand(addStudentCmd)
	courseCmd.AddCommand(rmStudentCmd)
	courseCmd.AddCommand(issueCourseCredentialCmd)
	rootCmd.AddCommand(courseCmd)
}
