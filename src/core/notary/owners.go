package notary

//go:generate abigen --abi ../../ethereum/build/abi/Owners.abi --bin ../../ethereum/build/bin/Owners.bin --pkg contract --type Owners --out ./contract/owners.go

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/relab/ct-eth-dapp/src/core/notary/contract"
)

var OwnersParams = &Params{
	ContractCode: contract.OwnersBin,
	ContractABI:  contract.OwnersABI,
}

// Owners is a Go wrapper around an on-chain owners contract.
type Owners struct {
	contract *contract.Owners
}

// NewOwners creates a struct exposing convenient operations to
// interact with the Owners contract.
func NewOwners(contractAddr common.Address, backend bind.ContractBackend) (*Owners, error) {
	c, err := contract.NewOwners(contractAddr, backend)
	if err != nil {
		return nil, err
	}
	return &Owners{contract: c}, nil
}

// IsOwner check if a given address is an Owner
func (c *Owners) IsOwner(opts *bind.CallOpts, address common.Address) (bool, error) {
	return c.contract.IsOwner(opts, address)
}

// Owners returns the list of owners
func (c *Owners) Owners(opts *bind.CallOpts) ([]common.Address, error) {
	return c.contract.Owners(opts)
}

// ChangeOwner one of the owners. Sender should be the old owner.
func (c *Owners) ChangeOwner(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return c.contract.ChangeOwner(opts, newOwner)
}
