package notary

//TODO: Use abi and bin files for binding and make manual deployment and link in the application, using truffle or using go directly, in both cases will be required to parse the json artifacts created during contract's compilation/deployment and retrieve the linking/address information.

//go:generate abigen --abi ../../ethereum/build/abi/CredentialSum.abi --bin ../../ethereum/build/bin/CredentialSum.bin --pkg contract --type CredentialSum --out ./contract/credential_sum.go

//go:generate abigen --abi ../../ethereum/build/abi/IssuerInterface.abi --bin ../../ethereum/build/bin/IssuerInterface.bin --pkg contract --type IssuerInterface --out ./contract/issuer_interface.go

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Params struct {
	ContractCode, ContractABI string
}

type Notary interface {
	RegisterCredential(opts *bind.TransactOpts, subject common.Address, digest [32]byte) (*types.Transaction, error)
	ConfirmCredential(opts *bind.TransactOpts, digest [32]byte) (*types.Transaction, error)
	Certified(opts *bind.CallOpts, digest [32]byte) (bool, error)
	Revoke(opts *bind.TransactOpts, digest [32]byte, reason [32]byte) (*types.Transaction, error)
	AggregateCredentials(opts *bind.TransactOpts, subject common.Address) (*types.Transaction, error)
	VerifyCredential(opts *bind.CallOpts, subject common.Address, digest [32]byte) (bool, error)
}
