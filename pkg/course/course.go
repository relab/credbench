package course

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	bindings "github.com/relab/bbchain-bindings/course"
	"github.com/relab/ct-eth-dapp/pkg/ctree/node"
	"github.com/relab/ct-eth-dapp/pkg/deployer"
)

// Course is a Go wrapper around an on-chain course contract.
type Course struct {
	*node.Node
	address  common.Address
	contract *bindings.Course
}

// NewCourse creates a struct exposing convenient operations to
// interact with the Course contract.
func NewCourse(contractAddr common.Address, backend bind.ContractBackend) (*Course, error) {
	cc, err := bindings.NewCourse(contractAddr, backend)
	if err != nil {
		return nil, err
	}
	n, err := node.NewNode(contractAddr, backend)
	if err != nil {
		return nil, err
	}
	return &Course{n, contractAddr, cc}, nil
}

func DeployCourse(auth *bind.TransactOpts, backend bind.ContractBackend, libs map[string]string, owners []common.Address, quorum uint8) (common.Address, *types.Transaction, *Course, error) {
	contractBin := deployer.LinkContract(bindings.CourseBin, libs)

	address, tx, _, err := deployer.DeployContract(auth, backend, bindings.CourseABI, contractBin, owners, quorum)
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	c, err := NewCourse(address, backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, c, nil
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

func (c *Course) GetCredentialProof(opts *bind.CallOpts, digest [32]byte) (bindings.NotaryCredentialProof, error) {
	return c.contract.GetCredentialProof(opts, digest)
}
