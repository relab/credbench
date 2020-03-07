package cmd

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/r0qs/bbchain-dapp/src/core/course"
	contract "github.com/r0qs/bbchain-dapp/src/core/go-bindings/course"
	"github.com/spf13/cobra"
)

var (
	courseContract *contract.Course
	courseAddress  string
)

var courseCmd = &cobra.Command{
	Use:               "course",
	Short:             "Course contract",
	PersistentPreRunE: clientPreRunE,
	PersistentPostRun: clientPostRun,
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

			backend, err := clientConn.Backend()
			senderTxOpts, err := wallet.GetTxOpts(backend)

			now := time.Now()
			startingTime := now.Unix()
			endingTime := now.Add(time.Hour).Unix()

			var cAddr common.Address
			cAddr, _, courseContract, err = contract.DeployCourse(senderTxOpts, backend, ownersList, big.NewInt(quorum), big.NewInt(startingTime), big.NewInt(endingTime))
			if err != nil {
				fmt.Printf("Failed on contract deployment: %v\n", err)
			}
			courseAddress = cAddr.Hex()
			fmt.Printf("Contract %v successfully deployed\n", cAddr.Hex())
		},
	}

	c.Flags().StringSliceVar(&owners, "owners", []string{}, "Owners addresses (comma separated)")
	c.Flags().Int64Var(&quorum, "quorum", int64(len(owners)), "Minimum number of signatures required to issue course credentials")

	c.MarkFlagRequired("owners")
	c.MarkFlagRequired("quorum")

	return c
}

var addStudentCmd = &cobra.Command{
	Use:   "addStudent",
	Short: "Add a student to the course contract",
	RunE: func(cmd *cobra.Command, args []string) error {
		studentAddress := common.HexToAddress(args[0])
		cAddr := common.HexToAddress(args[1])

		backend, err := clientConn.Backend()
		senderTxOpts, err := wallet.GetTxOpts(backend)

		course, err := course.NewCourse(cAddr, backend)
		if err != nil {
			return fmt.Errorf("Failed to get course: %v", err)
		}

		course.AddStudent(senderTxOpts, studentAddress)

		if ok, _ := course.IsEnrolled(&bind.CallOpts{Pending: true}, studentAddress); ok {
			fmt.Printf("student %s successfully enrolled!", studentAddress.Hex())
		}
		return nil
	},
}

func init() {
	courseCmd.PersistentFlags().StringVar(&courseAddress, "courseAddress", "", "Use specified course address")

	courseCmd.AddCommand(deployCourseCmd())
	courseCmd.AddCommand(addStudentCmd)
	rootCmd.AddCommand(courseCmd)
}
