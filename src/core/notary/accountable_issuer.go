package notary

//go:generate abigen --abi ../../ethereum/build/abi/AccountableIssuer.abi --bin ../../ethereum/build/bin/AccountableIssuer.bin --pkg aissuer --type AccountableIssuer --out ./contract/aissuer/accountable_issuer.go

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	contract "github.com/relab/ct-eth-dapp/src/core/notary/contract/aissuer"
)

var AccountableIssuerParams = &Params{ContractABI: contract.AccountableIssuerABI}

// AccountableIssuer is a Go wrapper around an on-chain AccountableIssuer contract.
type AccountableIssuer struct {
	*Issuer
	address  common.Address
	contract *contract.AccountableIssuer
}

// NewAccountableIssuer creates a struct exposing convenient operations to
// interact with the AccountableIssuer contract.
func NewAccountableIssuer(contractAddr common.Address, backend bind.ContractBackend) (*AccountableIssuer, error) {
	a, err := contract.NewAccountableIssuer(contractAddr, backend)
	if err != nil {
		return nil, err
	}
	i, err := NewIssuer(contractAddr, backend)
	return &AccountableIssuer{i, contractAddr, a}, nil
}

// Address returns the contract address of the issuer.
func (a *AccountableIssuer) Address() common.Address {
	return a.address
}

// OwnersList return the list of owners
func (a *AccountableIssuer) OwnersList(opts *bind.CallOpts) ([]common.Address, error) {
	return a.contract.Owners(opts)
}

// Issuers returns the list of registered issuers
func (a *AccountableIssuer) Issuers(opts *bind.CallOpts) ([]common.Address, error) {
	return a.contract.Issuers(opts)
}

// IsIssuer checks if a given contract address is a registered issuer
func (a *AccountableIssuer) IsIssuer(opts *bind.CallOpts, issuerAddress common.Address) (bool, error) {
	return a.contract.IsIssuer(opts, issuerAddress)
}

// AddIssuer registers a new issuer contract
func (a *AccountableIssuer) AddIssuer(opts *bind.TransactOpts, issuerAddress common.Address) (*types.Transaction, error) {
	return a.contract.AddIssuer(opts, issuerAddress)
}

// RegisterRootCredential collects all subject's credentials and issue a
// new credential proof iff the aggregation of those credentials on
// the sub-contracts match the given root (i.e. off-chain aggregation == on-chain aggregation)
func (a *AccountableIssuer) RegisterRootCredential(opts *bind.TransactOpts, subject common.Address, digest [32]byte, issuersAddresses []common.Address) (*types.Transaction, error) {
	return a.contract.RegisterCredential(opts, subject, digest, issuersAddresses)
}

// RegisterCredential issues a new credential proof ensuring append-only property
func (a *AccountableIssuer) RegisterCredential(opts *bind.TransactOpts, subject common.Address, digest [32]byte) (*types.Transaction, error) {
	return a.contract.RegisterCredential0(opts, subject, digest)
}

// VerifyCredentialTree performs a pre-order tree traversal over
// the credential tree of a given subject and verifies if the given
// root match with the current root on the root of the credential tree
// and if all the sub-trees were correctly built.
func (a *AccountableIssuer) VerifyCredentialTree(opts *bind.CallOpts, subject common.Address, root [32]byte) (bool, error) {
	return a.contract.VerifyCredentialTree(opts, subject, root)
}
