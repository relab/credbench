package faculty

//go:generate abigen --combined-json ../../ethereum/build/faculty/combined.json --pkg contract --out ./contract/faculty.go

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/r0qs/bbchain-dapp/src/core/faculty/contract"
	"github.com/r0qs/bbchain-dapp/src/core/notary/owners"
)

type Params struct {
	ContractCode, ContractAbi string
}

var ContractParams = &Params{contract.FacultyBin, contract.FacultyABI}

// Faculty is a Go wrapper around an on-chain faculty contract.
type Faculty struct {
	*owners.Owners
	address  common.Address
	contract *contract.Faculty
}

// NewFaculty creates a struct exposing convenient operations to
// interact with the Faculty contract.
func NewFaculty(contractAddr common.Address, backend bind.ContractBackend) (*Faculty, error) {
	c, err := contract.NewFaculty(contractAddr, backend)
	if err != nil {
		return nil, err
	}
	o, err := owners.NewOwners(contractAddr, backend)
	return &Faculty{o, contractAddr, c}, nil
}

// Address returns the on-chain contract address of the faculty.
func (c *Faculty) Address() common.Address {
	return c.address
}
