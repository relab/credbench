package course

//go:generate abigen --combined-json ../../ethereum/build/combined.json --pkg contract --out ../go-bindings/course/course.go

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/r0qs/bbchain-dapp/src/core/go-bindings/course"
	// "github.com/r0qs/bbchain-dapp/src/core/issuer"
)

type Params struct {
	ContractCode, ContractAbi string
}

var ContractParams = &Params{contract.CourseBin, contract.CourseABI}

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

func (c *Course) AddStudent(opts *bind.TransactOpts, student common.Address) (*types.Transaction, error) {
	return c.contract.AddStudent(opts, student)
}

func (c *Course) RemoveStudent(opts *bind.TransactOpts, student common.Address) (*types.Transaction, error) {
	return c.contract.RemoveStudent(opts, student)
}

func (c *Course) RenounceCourse(opts *bind.TransactOpts) (*types.Transaction, error) {
	return c.contract.RenounceCourse(opts)
}

func (c *Course) EnrolledStudents(opts *bind.CallOpts, student common.Address) (bool, error) {
	return c.contract.EnrolledStudents(opts, student)
}

func (c *Course) IsEnrolled(opts *bind.CallOpts, student common.Address) (bool, error) {
	return c.contract.IsEnrolled(opts, student)
}

// Owners functions
func (c *Course) IsOwner(address common.Address) (bool, error) {
	return c.contract.IsOwner(address)
}

func (c *Course) Owners() ([]common.Address, error) {
	length, err := c.contract.OwnersLength()
	var owners []common.Address
	i := big.NewInt(0)
	for i.Cmp(length) < 0 {
		owner, _ := c.contract.Owners(i)
		owners = append(owners, owner)
		i.Add(i, big.NewInt(1))
	}
	return owners, err
}

// Issuer functions
func (c *Course) Issue(student common.Address, digest [32]byte) (*types.Transaction, error) {
	transactOpts := bind.NewKeyedTransactor(c.prvKey)
	return c.contract.Issue(transactOpts, student, digest)
}

func (c *Course) Revoke(digest [32]byte) (*types.Transaction, error) {
	transactOpts := bind.NewKeyedTransactor(c.prvKey)
	return c.contract.Revoke(transactOpts, digest)
}

func (c *Course) Issued(digest [32]byte) *issuer.CredentialProof {
	proof, _ := c.contract.Issued(digest)
	var cp issuer.CredentialProof = proof
	return &cp
}

func (c *Course) Revoked(digest [32]byte) *issuer.RevokeProof {
	proof, _ := c.contract.Revoked(digest)
	var rp issuer.RevokeProof = proof
	return &rp
}
