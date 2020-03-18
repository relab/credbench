package owners

//go:generate abigen --abi ../../../ethereum/build/abi/Owners.abi --bin ../../../ethereum/build/bin/Owners.bin --pkg contract --type Owners --out ./contract/owners.go

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/r0qs/bbchain-dapp/src/core/notary/owners/contract"
	"math/big"
)

type Params struct {
	ContractCode, ContractAbi string
}

var ContractParams = &Params{contract.OwnersBin, contract.OwnersABI}

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

// IsOwner
func (c *Owners) IsOwner(opts *bind.CallOpts, address common.Address) (bool, error) {
	return c.contract.IsOwner(opts, address)
}

// OwnersList
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