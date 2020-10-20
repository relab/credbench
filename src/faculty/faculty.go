package faculty

//go:generate abigen --abi ../ethereum/build/abi/Faculty.abi --bin ../ethereum/build/bin/Faculty.bin --pkg faculty --type FacultyContract --out ./faculty_contract.go

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/relab/ct-eth-dapp/src/ctree/node"
	"github.com/relab/ct-eth-dapp/src/ctree/notary/issuer"
	"github.com/relab/ct-eth-dapp/src/deployer"
)

// Faculty is a Go wrapper around an on-chain faculty contract.
type Faculty struct {
	*node.Node
	*issuer.Issuer
	address  common.Address
	contract *FacultyContract
}

// NewFaculty creates a struct exposing convenient operations to
// interact with the Faculty contract.
func NewFaculty(contractAddr common.Address, backend bind.ContractBackend) (*Faculty, error) {
	cc, err := NewFacultyContract(contractAddr, backend)
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
	return &Faculty{n, i, contractAddr, cc}, nil
}

func DeployFaculty(auth *bind.TransactOpts, backend bind.ContractBackend, libs map[string]string, owners []common.Address, quorum uint8) (common.Address, *types.Transaction, *Faculty, error) {
	contractBin := deployer.LinkContract(FacultyContractBin, libs)

	address, tx, contract, err := deployer.DeployContract(auth, backend, FacultyContractABI, contractBin, owners, quorum)
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	c := &FacultyContract{FacultyContractCaller: FacultyContractCaller{contract: contract}, FacultyContractTransactor: FacultyContractTransactor{contract: contract}, FacultyContractFilterer: FacultyContractFilterer{contract: contract}}
	return address, tx, &Faculty{address: address, contract: c}, nil
}

// Address returns the contract address of the faculty.
func (f Faculty) Address() common.Address {
	return f.address
}
