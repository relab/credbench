package course

//go:generate abigen --sol ../../ethereum/contracts/Course.sol --exc ../../ethereum/contracts/Issuer.sol:Issuer,../../ethereum/contracts/Owners.sol:Owners --pkg contract --out ../go-bindings/course/course.go

//abigen --combined-json ../../ethereum/build/combined.json --pkg contract --out ../go-bindings/course/course.go

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/r0qs/dvcp/src/core/go-bindings/course"
	"github.com/r0qs/dvcp/src/core/issuer"
)

type Params struct {
	ContractCode, ContractAbi string
}

var ContractParams = &Params{contract.CourseBin, contract.CourseABI}

type Course struct {
	prvKey       *ecdsa.PrivateKey
	contractAddr common.Address
	session      *contract.CourseSession
	backend      bind.ContractBackend
}

// DeployCourse deploys an instance of the Course contract
func DeployCourse(transactOpts *bind.TransactOpts, backend bind.ContractBackend, owners []common.Address, quorum *big.Int) (common.Address, error) {

	courseAddr, _, _, err := contract.DeployCourse(transactOpts, backend, owners, quorum)
	if err != nil {
		return common.Address{}, err
	}
	return courseAddr, nil
}

// NewCourse creates a struct exposing convenient operations to
// interact with the Course contract.
func NewCourse(transactOpts *bind.TransactOpts, backend bind.ContractBackend, contractAddr common.Address, prvKey *ecdsa.PrivateKey) (*Course, error) {

	contractCourse, err := contract.NewCourse(contractAddr, backend)
	if err != nil {
		return nil, err
	}

	return &Course{
		prvKey,
		contractAddr,
		&contract.CourseSession{
			Contract:     contractCourse,
			TransactOpts: *transactOpts,
		},
		backend,
	}, nil
}

// Address returns the on-chain contract address of the course.
func (c *Course) Address() common.Address {
	return c.contractAddr
}

func (c *Course) AddStudent(student common.Address) (*types.Transaction, error) {
	transactOpts := bind.NewKeyedTransactor(c.prvKey)
	return c.session.Contract.AddStudent(transactOpts, student)
}

func (c *Course) RemoveStudent(student common.Address) (*types.Transaction, error) {
	transactOpts := bind.NewKeyedTransactor(c.prvKey)
	return c.session.Contract.RemoveStudent(transactOpts, student)
}

func (c *Course) RenounceCourse() (*types.Transaction, error) {
	transactOpts := bind.NewKeyedTransactor(c.prvKey)
	return c.session.Contract.RenounceCourse(transactOpts)
}

func (c *Course) EnrolledStudents(student common.Address) (bool, error) {
	return c.session.EnrolledStudents(student)
}

func (c *Course) IsEnrolled(student common.Address) (bool, error) {
	return c.session.IsEnrolled(student)
}

// Owners functions
func (c *Course) IsOwner(address common.Address) (bool, error) {
	return c.session.IsOwner(address)
}

func (c *Course) Owners() ([]common.Address, error) {
	length, err := c.session.OwnersLength()
	var owners []common.Address
	i := big.NewInt(0)
	for i.Cmp(length) < 0 {
		owner, _ := c.session.Owners(i)
		owners = append(owners, owner)
		i.Add(i, big.NewInt(1))
	}
	return owners, err
}

// Issuer functions
func (c *Course) Issue(student common.Address, digest [32]byte) (*types.Transaction, error) {
	transactOpts := bind.NewKeyedTransactor(c.prvKey)
	return c.session.Contract.Issue(transactOpts, student, digest)
}

func (c *Course) Revoke(digest [32]byte) (*types.Transaction, error) {
	transactOpts := bind.NewKeyedTransactor(c.prvKey)
	return c.session.Contract.Revoke(transactOpts, digest)
}

func (c *Course) Issued(digest [32]byte) *issuer.CredentialProof {
	proof, _ := c.session.Issued(digest)
	var cp issuer.CredentialProof = proof
	return &cp
}

func (c *Course) Revoked(digest [32]byte) *issuer.RevokeProof {
	proof, _ := c.session.Revoked(digest)
	var rp issuer.RevokeProof = proof
	return &rp
}
