package faculty

//go:generate abigen --combined-json ../../ethereum/build/faculty/combined.json --pkg contract --out ./contract/faculty.go

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/relab/bbchain-dapp/src/core/faculty/contract"
	"github.com/relab/bbchain-dapp/src/core/notary"
)

var FacultyParams = &notary.Params{
	ContractCode: contract.FacultyBin,
	ContractABI:  contract.FacultyABI,
}

// Faculty is a Go wrapper around an on-chain faculty contract.
type Faculty struct {
	*notary.AccountableIssuer
	address  common.Address
	contract *contract.Faculty
}

// NewFaculty creates a struct exposing convenient operations to
// interact with the Faculty contract.
func NewFaculty(contractAddr common.Address, backend bind.ContractBackend) (*Faculty, error) {
	f, err := contract.NewFaculty(contractAddr, backend)
	if err != nil {
		return nil, err
	}
	a, err := notary.NewAccountableIssuer(contractAddr, backend)
	return &Faculty{a, contractAddr, f}, nil
}

// Address returns the contract address of the faculty.
func (f *Faculty) Address() common.Address {
	return f.address
}

// CoursesBySemester returns the list of courses for the given semester
func (f *Faculty) CoursesBySemester(opts *bind.CallOpts, semester [32]byte) []common.Address {
	var courses []common.Address
	for i := 0; ; i++ {
		c, err := f.contract.CoursesBySemester(opts, semester, big.NewInt(int64(i)))
		if err != nil {
			break
		}
		courses = append(courses, c)
	}
	return courses
}

// CreateCourse create a new course contract
func (f *Faculty) CreateCourse(opts *bind.TransactOpts, semester [32]byte, teachers []common.Address, quorum *big.Int) (*types.Transaction, error) {
	return f.contract.CreateCourse(opts, semester, teachers, quorum)
}
