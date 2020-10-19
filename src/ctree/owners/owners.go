package owners

//go:generate abigen --abi ../../ethereum/build/abi/Owners.abi --bin ../../ethereum/build/bin/Owners.bin --pkg owners --type OwnersContract --out ./owners_contract.go

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Owners is a Go wrapper around an owners contract.
type Owners struct {
	contract *OwnersContract
}

// NewOwners creates a struct exposing convenient operations to
// interact with the Owners contract.
func NewOwners(contractAddr common.Address, backend bind.ContractBackend) (*Owners, error) {
	c, err := NewOwnersContract(contractAddr, backend)
	if err != nil {
		return nil, err
	}
	return &Owners{contract: c}, nil
}

// IsOwner check if a given address is an Owner
func (c Owners) IsOwner(opts *bind.CallOpts, address common.Address) (bool, error) {
	return c.contract.IsOwner(opts, address)
}

// GetOwners returns the list of owners
func (c Owners) GetOwners(opts *bind.CallOpts) ([]common.Address, error) {
	return c.contract.Owners(opts)
}

// Quorum returns the list of owners
func (c Owners) Quorum(opts *bind.CallOpts) (uint8, error) {
	return c.contract.Quorum(opts)
}

// ChangeOwner one of the owners. Sender should be the old owner.
func (c *Owners) ChangeOwner(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return c.contract.ChangeOwner(opts, newOwner)
}
