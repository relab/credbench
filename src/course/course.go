package course

//go:generate abigen --sol ../ethereum/contracts/Course.sol --pkg contract --out ../go-bindings/course/course.go

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/r0qs/dvcp/src/go-bindings/course"
)

type Params struct {
	ContractCode, ContractAbi string
}

var ContractParams = &Params{contract.CourseBin, contract.CourseABI}

type Course struct {
	contract        *contract.Course // abigen binding
	contractAddr    common.Address   // contract address
	contractBackend bind.ContractBackend
}

// DeployCourse deploys an instance of the Course contract
func DeployCourse(transactOpts *bind.TransactOpts, contractBackend bind.ContractBackend, owners []common.Address, quorum *big.Int) (common.Address, *Course, error) {

	courseAddr, _, _, err := contract.DeployCourse(transactOpts, contractBackend, owners, quorum)
	if err != nil {
		return courseAddr, nil, err
	}

	course, err := NewCourse(courseAddr, contractBackend)
	if err != nil {
		return courseAddr, nil, err
	}

	return courseAddr, course, nil
}

// NewCourse creates a struct exposing convenient high-level operations to
// interact with the Course contract.
func NewCourse(contractAddr common.Address, contractBackend bind.ContractBackend) (*Course, error) {

	contractCourse, err := contract.NewCourse(contractAddr, contractBackend)
	if err != nil {
		return nil, err
	}

	return &Course{
		contract:        contractCourse,
		contractAddr:    contractAddr,
		contractBackend: contractBackend,
	}, nil
}

// Owners returns the array of owners of the course.
// TODO: Look for a better way to do this and create a
// class for owners and notary (until now the contracts are deployed together)
func (c *Course) Owners() ([]common.Address, error) {
	length, err := c.contract.OwnersLength(nil)
	var owners []common.Address
	i := big.NewInt(0)
	for i.Cmp(length) < 0 {
		owner, _ := c.contract.AllOwners(nil, i)
		owners = append(owners, owner)
		i.Add(i, big.NewInt(1))
	}
	return owners, err
}

// Address returns the on-chain contract address of the course.
func (c *Course) Address() common.Address {
	return c.contractAddr
}

// Contract returns the instance of the Course.
func (c *Course) Contract() *contract.Course {
	return c.contract
}
