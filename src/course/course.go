package course

//go:generate abigen --abi ../ethereum/build/abi/Course.abi --bin ../ethereum/build/bin/Course.bin --pkg course --type CourseContract --out ./course_contract.go

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/relab/ct-eth-dapp/src/ctree/node"
	"github.com/relab/ct-eth-dapp/src/ctree/notary/issuer"
	"github.com/relab/ct-eth-dapp/src/deployer"
)

// Course is a Go wrapper around an on-chain course contract.
type Course struct {
	*node.Node
	*issuer.Issuer
	address  common.Address
	contract *CourseContract
}

// NewCourse creates a struct exposing convenient operations to
// interact with the Course contract.
func NewCourse(contractAddr common.Address, backend bind.ContractBackend) (*Course, error) {
	cc, err := NewCourseContract(contractAddr, backend)
	if err != nil {
		return nil, err
	}
	n, err := node.NewNode(contractAddr, backend)
	if err != nil {
		return nil, err
	}
	i, err := issuer.NewIssuer(contractAddr, backend)
	if err != nil {
		return nil, err
	}
	return &Course{n, i, contractAddr, cc}, nil
}

func DeployCourse(auth *bind.TransactOpts, backend bind.ContractBackend, libs map[string]string, owners []common.Address, quorum uint8) (common.Address, *types.Transaction, *Course, error) {
	contractBin := deployer.LinkContract(CourseContractBin, libs)

	address, tx, contract, err := deployer.DeployContract(auth, backend, CourseContractABI, contractBin, owners, quorum)
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	c := &CourseContract{CourseContractCaller: CourseContractCaller{contract: contract}, CourseContractTransactor: CourseContractTransactor{contract: contract}, CourseContractFilterer: CourseContractFilterer{contract: contract}}
	return address, tx, &Course{address: address, contract: c}, nil
}

// Address returns the contract address of the course.
func (c Course) Address() common.Address {
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

// GetStudents
func (c Course) GetStudents(opts *bind.CallOpts) ([]common.Address, error) {
	return c.contract.GetStudents(opts)
}

// IsEnrolled
func (c *Course) IsEnrolled(opts *bind.CallOpts, student common.Address) (bool, error) {
	return c.contract.IsEnrolled(opts, student)
}

func (c *Course) RegisterExam(opts *bind.TransactOpts, student common.Address, examDigest [32]byte) (*types.Transaction, error) {
	return c.contract.RegisterExam(opts, student, examDigest)
}
