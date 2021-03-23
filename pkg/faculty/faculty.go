package faculty

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	bindings "github.com/relab/bbchain-bindings/faculty"
	"github.com/relab/ct-eth-dapp/pkg/ctree/node"
	"github.com/relab/ct-eth-dapp/pkg/deployer"
)

// Faculty is a Go wrapper around an on-chain faculty contract.
type Faculty struct {
	*node.Node
	address  common.Address
	contract *bindings.Faculty
}

// NewFaculty creates a struct exposing convenient operations to
// interact with the Faculty contract.
func NewFaculty(contractAddr common.Address, backend bind.ContractBackend) (*Faculty, error) {
	cc, err := bindings.NewFaculty(contractAddr, backend)
	if err != nil {
		return nil, err
	}
	n, err := node.NewNode(contractAddr, backend)
	if err != nil {
		return nil, err
	}
	return &Faculty{n, contractAddr, cc}, nil
}

func DeployFaculty(auth *bind.TransactOpts, backend bind.ContractBackend, libs map[string]string, owners []common.Address, quorum uint8) (common.Address, *types.Transaction, *Faculty, error) {
	contractBin := deployer.LinkContract(bindings.FacultyBin, libs)

	address, tx, _, err := deployer.DeployContract(auth, backend, bindings.FacultyABI, contractBin, owners, quorum)
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	f, err := NewFaculty(address, backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, f, nil
}

// Address returns the contract address of the faculty.
func (f Faculty) Address() common.Address {
	return f.address
}

func (f Faculty) SemesterExists(opts *bind.CallOpts, semester [32]byte) (bool, error) {
	return f.contract.SemesterExists(opts, semester)
}

func (f Faculty) GetCoursesBySemester(opts *bind.CallOpts, semester []byte) ([]common.Address, error) {
	var s [32]byte
	copy(s[:], semester[:]) // truncate to 32 bytes
	return f.contract.GetCoursesBySemester(opts, s)
}

func (f *Faculty) RegisterSemester(opts *bind.TransactOpts, semester [32]byte, courses []common.Address) (*types.Transaction, error) {
	return f.contract.RegisterSemester(opts, semester, courses)
}
