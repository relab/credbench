package cmd

import (
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/relab/bbchain-dapp/src/core/course"
	"github.com/spf13/cobra"

	contract "github.com/relab/bbchain-dapp/src/core/course/contract"
	pb "github.com/relab/bbchain-dapp/src/schemes"
)

var (
	courseContract *contract.Course
	courseAddress  string
)

var courseCmd = &cobra.Command{
	Use:   "course",
	Short: "Course contract",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		err := setupClient()
		if err != nil {
			log.Fatalln(err.Error())
		}
	},
	PersistentPostRun: clientClose,
}

func deployCourseCmd() *cobra.Command {
	var owners []string
	var quorum int64

	c := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy course contract",
		Run: func(cmd *cobra.Command, args []string) {
			var ownersList []common.Address
			for _, addr := range owners {
				ownersList = append(ownersList, common.HexToAddress(addr))
			}
			tx, err := deployCourse(ownersList, quorum)
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

func deployCourse(owners []common.Address, quorum int64) (tx *types.Transaction, err error) {
	backend, _ := clientConn.Backend()
	opts, _ := wallet.GetTxOpts(backend)

	now := time.Now()
	startingTime := now.Unix()
	endingTime := now.Add(time.Hour).Unix()

	var cAddr common.Address
	cAddr, tx, courseContract, err = contract.DeployCourse(opts, backend, owners, big.NewInt(quorum), big.NewInt(startingTime), big.NewInt(endingTime))
	courseAddress = cAddr.Hex()
	if err != nil || courseAddress == "0x0000000000000000000000000000000000000000" {
		return nil, fmt.Errorf("failed to deploy the contract: %v", err)
	}
	fmt.Printf("Contract %v successfully deployed\n", courseAddress)
	return tx, nil
}

var addStudentCmd = &cobra.Command{
	Use:   "addStudent",
	Short: "Add a student to the course contract",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		studentAddress := common.HexToAddress(args[0])
		err := addStudent(studentAddress)
		if err != nil {
			log.Fatalln(err.Error())
		}
	},
}

func addStudent(studentAddress common.Address) error {
	cAddr := common.HexToAddress(courseAddress)
	backend, _ := clientConn.Backend()
	opts, _ := wallet.GetTxOpts(backend)

	course, err := course.NewCourse(cAddr, backend)
	if err != nil {
		return fmt.Errorf("Failed to get course: %v", err)
	}

	_, err = course.AddStudent(opts, studentAddress)
	if err != nil {
		return fmt.Errorf("Failed to add student: %v", err)
	}

	if ok, _ := course.IsEnrolled(&bind.CallOpts{Pending: true}, studentAddress); ok {
		fmt.Printf("student %s successfully enrolled!", studentAddress.Hex())
	}
	return nil
}

var rmStudentCmd = &cobra.Command{
	Use:   "rmStudent",
	Short: "Remove a student to the course contract",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		studentAddress := common.HexToAddress(args[0])
		err := rmStudent(studentAddress)
		if err != nil {
			log.Fatalln(err.Error())
		}

	},
}

func rmStudent(studentAddress common.Address) error {
	cAddr := common.HexToAddress(courseAddress)
	backend, _ := clientConn.Backend()
	opts, _ := wallet.GetTxOpts(backend)

	course, err := course.NewCourse(cAddr, backend)
	if err != nil {
		return fmt.Errorf("Failed to get course: %v", err)
	}

	_, err = course.RemoveStudent(opts, studentAddress)
	if err != nil {
		return fmt.Errorf("Failed to remove student: %v", err)
	}

	if ok, _ := course.IsEnrolled(nil, studentAddress); ok {
		fmt.Printf("student %s successfully removed!", studentAddress.Hex())
	}
	return nil
}

var issueCourseCredentialCmd = &cobra.Command{
	Use:   "issue",
	Short: "Issue a new credential",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("Missing arguments. Please specify: student_address path_to_json_credential")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		studentAddress := common.HexToAddress(args[0])
		a := &pb.AssignmentGradeCredential{}
		pb.ParseJSON(args[1], a)
		digest := pb.Hash(a)
		err := registerCredential(studentAddress, digest)
		if err != nil {
			log.Fatalln(err.Error())
		}
	},
}

func registerCredential(studentAddress common.Address, digest [32]byte) error {
	cAddr := common.HexToAddress(courseAddress)
	backend, err := clientConn.Backend()
	opts, err := wallet.GetTxOpts(backend)

	course, err := course.NewCourse(cAddr, backend)
	if err != nil {
		return fmt.Errorf("Failed to get course: %v", err)
	}

	_, err = course.RegisterCredential(opts, studentAddress, digest)
	if err != nil {
		return fmt.Errorf("Failed to register credential: %v", err)
	}
	return nil
}

func init() {
	courseCmd.PersistentFlags().StringVar(&courseAddress, "courseAddress", "", "Use specified course address")

	courseCmd.AddCommand(deployCourseCmd())
	courseCmd.AddCommand(addStudentCmd)
	courseCmd.AddCommand(rmStudentCmd)
	courseCmd.AddCommand(issueCourseCredentialCmd)
	rootCmd.AddCommand(courseCmd)
}
