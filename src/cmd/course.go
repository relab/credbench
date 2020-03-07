package cmd

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/r0qs/bbchain-dapp/src/core/course"
	"github.com/spf13/cobra"
)

var (
	courseAddress string
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

			cAddr, err := course.DeployCourse(senderTxOpts, backend, ownersList, big.NewInt(quorum))
			if err != nil {
				fmt.Printf("Failed on contract deployment: %v\n", err)
			}
			fmt.Println(cAddr.Hex())
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

		course, err := course.NewCourse(senderTxOpts, backend, cAddr, wallet.PrivateKey())
		if err != nil {
			return fmt.Errorf("Failed to get course: %v", err)
		}

		course.AddStudent(studentAddress)

		ok, _ := course.IsEnrolled(studentAddress)
		if ok {
			fmt.Printf("student %s successfully enrolled!", studentAddress.Hex())
		}
		return nil
	},
}

func init() {
	courseCmd.PersistentFlags().StringVar(&courseAddress, "courseAddress", "", "Use specified course address")

	rootCmd.AddCommand(courseCmd)
	courseCmd.AddCommand(deployCourseCmd())
	courseCmd.AddCommand(addStudentCmd)
}
