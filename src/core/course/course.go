package course

//go:generate abigen --combined-json ../../ethereum/build/combined.json --pkg contract --out ./contract/course.go

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/r0qs/bbchain-dapp/src/core/course/contract"
	"github.com/r0qs/bbchain-dapp/src/core/notary"
	"math/big"
)

type Params struct {
	ContractCode, ContractAbi string
}

var ContractParams = &Params{contract.CourseBin, contract.CourseABI}

// Course is a Go wrapper around an on-chain course contract.
type Course struct {
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
	return &Course{address: contractAddr, contract: c}, nil
}

// Address returns the on-chain contract address of the course.
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

// IsOwner
func (c *Course) IsOwner(opts *bind.CallOpts, address common.Address) (bool, error) {
	return c.contract.IsOwner(opts, address)
}

// Owners
func (c *Course) Owners(opts *bind.CallOpts) ([]common.Address, error) {
	length, err := c.contract.OwnersLength(opts)
	var owners []common.Address
	i := big.NewInt(0)
	for i.Cmp(length) < 0 {
		owner, _ := c.contract.Owners(opts, i)
		owners = append(owners, owner)
		i.Add(i, big.NewInt(1))
	}
	return owners, err
}

// RegisterCredential
func (c *Course) RegisterCredential(opts *bind.TransactOpts, student common.Address, digest [32]byte) (*types.Transaction, error) {
	return c.contract.RegisterCredential(opts, student, digest)
}

// ConfirmCredential
func (c *Course) ConfirmCredential(opts *bind.TransactOpts, digest [32]byte) (*types.Transaction, error) {
	return c.contract.ConfirmCredential(opts, digest)
}

// Revoke
func (c *Course) Revoke(opts *bind.TransactOpts, digest [32]byte, reason [32]byte) (*types.Transaction, error) {
	return c.contract.RevokeCredential(opts, digest, reason)
}

func (c *Course) DigestsBySubject(opts *bind.CallOpts, subject common.Address) ([][32]byte, error) {
	return c.contract.DigestsBySubject(opts, subject)
}

// IssuedCredentials
func (c *Course) IssuedCredentials(opts *bind.CallOpts, digest [32]byte) *notary.CredentialProof {
	proof, _ := c.contract.IssuedCredentials(opts, digest)
	var cp notary.CredentialProof = proof
	return &cp
}

// RevokedCredentials
func (c *Course) RevokedCredentials(opts *bind.CallOpts, digest [32]byte) *notary.RevocationProof {
	proof, _ := c.contract.RevokedCredentials(opts, digest)
	var rp notary.RevocationProof = proof
	return &rp
}
