package cmd

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	bindings "github.com/relab/bbchain-bindings/faculty"
	"github.com/relab/ct-eth-dapp/bench/datastore"
	"github.com/relab/ct-eth-dapp/bench/transactor"
	faculty "github.com/relab/ct-eth-dapp/pkg/faculty"
)

func registerSemesterCredential(e *transactor.Transactor, opts *bind.TransactOpts, f *faculty.Faculty, studentAddress common.Address, digest [32]byte, witnesses []common.Address) (*types.Transaction, error) {
	tx, err := e.SendTX("faculty", opts, f.Address(), bindings.FacultyABI, "registerCredential", studentAddress, digest, witnesses)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func approveSemesterCredential(e *transactor.Transactor, opts *bind.TransactOpts, f *faculty.Faculty, digest [32]byte) (*types.Transaction, error) {
	tx, err := e.SendTX("faculty", opts, f.Address(), bindings.FacultyABI, "approveCredential", digest)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func aggregateSemesterCredentials(e *transactor.Transactor, opts *bind.TransactOpts, f *faculty.Faculty, student common.Address, digests [][32]byte) (*types.Transaction, error) {
	tx, err := e.SendTX("faculty", opts, f.Address(), bindings.FacultyABI, "aggregateCredentials", student, digests)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func getFacultyContract(facultyAddress common.Address) (*faculty.Faculty, error) {
	c, err := faculty.NewFaculty(facultyAddress, backend)
	if err != nil {
		return nil, fmt.Errorf("Failed to get faculty: %v", err)
	}
	return c, nil
}

var getFacultyCmd = &cobra.Command{
	Use:   "get",
	Short: "Shows the faculty details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		address := common.HexToAddress(args[0])
		fs := datastore.NewFacultyStore(db, address)
		faculty, err := fs.GetFaculty()
		if err != nil {
			log.Error(err)
		}
		fmt.Printf("Faculty Info:\n")
		fmt.Printf("\tAddress: %s\n", address.Hex())
		fmt.Printf("\tCreated on: %s\n", faculty.CreatedOn.AsTime())
		fmt.Printf("\tAdministrators:\n")
		for _, e := range faculty.Adms {
			fmt.Printf("\t  %s\n", common.BytesToAddress(e).Hex())
		}
		fmt.Printf("\tSemesters:\n")
		for _, s := range faculty.Semesters {
			fmt.Printf("\t  %s\n", common.Bytes2Hex(s))
		}
		fmt.Printf("\tStudents:\n")
		for s := range faculty.Students {
			fmt.Printf("\t  %s\n", s)
		}
	},
}

func newFacultyCmd() *cobra.Command {
	facultyCmd := &cobra.Command{
		Use:   "faculty",
		Short: "Manage faculty",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			rootCmd.PersistentPreRun(cmd, args)
			err := loadDefaultAccount()
			if err != nil {
				log.Fatal(err)
			}
		},
	}
	facultyCmd.AddCommand(
		getFacultyCmd,
	)
	return facultyCmd
}
