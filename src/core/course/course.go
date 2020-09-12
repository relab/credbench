package course

//go:generate abigen --combined-json ../../ethereum/build/course/combined.json --pkg contract --out ./contract/course.go

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/relab/ct-eth-dapp/src/core/course/contract"
	"github.com/relab/ct-eth-dapp/src/core/notary"
)

var CourseParams = &notary.Params{
	ContractCode: contract.CourseBin,
	ContractABI:  contract.CourseABI,
}

// Course is a Go wrapper around an on-chain course contract.
type Course struct {
	*notary.Issuer
	address  common.Address
	contract *contract.Course
}

// NewCourse creates a struct exposing convenient operations to
// interact with the Course contract.
func NewCourse(contractAddr common.Address, backend bind.ContractBackend) (*Course, error) {
	c, err := contract.NewCourse(contractAddr, backend)
	if err != nil {
		return nil, err
	}
	i, err := notary.NewIssuer(contractAddr, backend)
	return &Course{i, contractAddr, c}, nil
}

// Address returns the contract address of the course.
func (c *Course) Address() common.Address {
	return c.address
}

// AddStudent
func (c *Course) AddStudent(opts *bind.TransactOpts, student common.Address) (*types.Transaction, error) {
	return c.contract.AddStudent(opts, student)
}

// RemoveStudent
func (c *Course) RemoveStudent(opts *bind.TransactOpts, student common.Address) (*types.Transaction, error) {
	return c.contract.RemoveStudent(opts, student)
}

// RenounceCourse
func (c *Course) RenounceCourse(opts *bind.TransactOpts) (*types.Transaction, error) {
	return c.contract.RenounceCourse(opts)
}

// EnrolledStudents
func (c *Course) EnrolledStudents(opts *bind.CallOpts, student common.Address) (bool, error) {
	return c.contract.EnrolledStudents(opts, student)
}

// IsEnrolled
func (c *Course) IsEnrolled(opts *bind.CallOpts, student common.Address) (bool, error) {
	return c.contract.IsEnrolled(opts, student)
}
