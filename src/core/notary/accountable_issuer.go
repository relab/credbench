package notary

//go:generate abigen --abi ../../ethereum/build/abi/AccountableIssuer.abi --bin ../../ethereum/build/bin/AccountableIssuer.bin --pkg contract --type AccountableIssuer --out ./contract/accountable_issuer.go

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/r0qs/bbchain-dapp/src/core/notary/contract"
)

var AccountableIssuerParams = &Params{ContractAbi: contract.OwnersABI}

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

// Issuers returns the list of registered issuers
func (a *AccountableIssuer) Issuers(opts *bind.CallOpts) ([]common.Address, error) {
	length, err := a.contract.IssuersLength(opts)
	var issuers []common.Address
	i := big.NewInt(0)
	for i.Cmp(length) < 0 {
		issuer, _ := a.contract.Issuers(opts, i)
		issuers = append(issuers, issuer)
		i.Add(i, big.NewInt(1))
	}
	return issuers, err
}

// IsIssuer checks if a given contract address is a registered issuer
func (a *AccountableIssuer) IsIssuer(opts *bind.CallOpts, issuerAddress common.Address) (bool, error) {
	return a.contract.IsIssuer(opts, issuerAddress)
}

// AddIssuer registers a new issuer contract
func (a *AccountableIssuer) AddIssuer(opts *bind.TransactOpts, issuerAddress common.Address) (*types.Transaction, error) {
	return a.contract.AddIssuer(opts, issuerAddress)
}

// CollectCredentials collects all the aggregated digests of
// a given subject on all given registered sub-contracts
func (a *AccountableIssuer) CollectCredentials(opts *bind.CallOpts, subject common.Address, issuersAddresses []common.Address) ([][32]byte, error) {
	return a.contract.CollectCredentials(opts, subject, issuersAddresses)
}

// RegisterRootCredential collects all subject's credentials and issue a
// new credential proof iff the aggregation of those credentials on
// the sub-contracts match the given root (i.e. off-chain aggregation == on-chain aggregation)
func (a *AccountableIssuer) RegisterRootCredential(opts *bind.TransactOpts, subject common.Address, digest [32]byte, digestRoot [32]byte, issuersAddresses []common.Address) (*types.Transaction, error) {
	return a.contract.RegisterCredential0(opts, subject, digest, digestRoot, issuersAddresses)
}

// RegisterCredential issues a new credential proof ensuring append-only property
func (a *AccountableIssuer) RegisterCredential(opts *bind.TransactOpts, subject common.Address, digest [32]byte) (*types.Transaction, error) {
	return a.contract.RegisterCredential(opts, subject, digest)
}

// VerifyCredential iteractivally verifies if a given credential proof
// (i.e. represented by it's digest) corresponds to the aggregation
// of all stored credentials of a particular subject in all given sub-contracts
func (a *AccountableIssuer) VerifyCredential(opts *bind.CallOpts, subject common.Address, proofs [][32]byte, issuersAddresses []common.Address) (bool, error) {
	return a.contract.VerifyCredential(opts, subject, proofs, issuersAddresses)
}
