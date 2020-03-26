package notary

//go:generate abigen --abi ../../ethereum/build/abi/Owners.abi --bin ../../ethereum/build/bin/Owners.bin --pkg contract --type Owners --out ./contract/owners.go

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/relab/bbchain-dapp/src/core/notary/contract"
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

// OwnersList returns the length of the owners array
func (c *Owners) OwnersList(opts *bind.CallOpts) ([]common.Address, error) {
	length, err := c.contract.OwnersLength(opts)
	var owners []common.Address
	i := big.NewInt(0)
	for i.Cmp(length) < 0 {
		owner, _ := c.contract.Owners(opts, i)
		owners = append(owners, owner)
		i.Add(i, big.NewInt(1))
	}
	return owners, err
}

// ChangeOwner one of the owners. Sender should be the old owner.
func (c *Owners) ChangeOwner(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return c.contract.ChangeOwner(opts, newOwner)
}
