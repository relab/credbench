package course

//go:generate abigen --sol ../ethereum/contracts/Course.sol --exc ../ethereum/contracts/Notary.sol:Notary,../ethereum/contracts/Owners.sol:Owners --pkg contract --out ../go-bindings/course/course.go

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/r0qs/dvcp/src/go-bindings/course"
	"math/big"
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
