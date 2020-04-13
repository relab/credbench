package notary

//TODO: Use abi and bin files for binding and make manual deployement and link in the application, using truffle or using go directly, in both cases will be required to parse the json artifacts created during contract's compilation/deployement and retrieve the linking/address information.

//go:generate abigen --abi ../../ethereum/build/abi/CredentialSum.abi --bin ../../ethereum/build/bin/CredentialSum.bin --pkg contract --type CredentialSum --out ./contract/credential_sum.go

//go:generate abigen --abi ../../ethereum/build/abi/IssuerInterface.abi --bin ../../ethereum/build/bin/IssuerInterface.bin --pkg contract --type IssuerInterface --out ./contract/issuer_interface.go

import (
	"fmt"
	"math/big"

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

// CredentialProof represents an on-chain proof that a
// verifiable credential was created and signed by an issuer.
type CredentialProof struct {
	Signed         *big.Int
	Approved       bool
	InsertedBlock  *big.Int
	BlockTimestamp *big.Int
	Nonce          *big.Int
	Issuer         common.Address
	Subject        common.Address
	Digest         [32]byte
}

func (cp *CredentialProof) String() string {
	return fmt.Sprintf("Signed: %v, Approved: %t, InsertedBlock: %v, Issuer: %s, Subject: %s, Digest: %x", cp.Signed, cp.Approved, cp.InsertedBlock, cp.Issuer.Hex(), cp.Subject.Hex(), cp.Digest)
}

// RevocationProof represents an on-chain proof that a
// verifiable credential was revoked by an issuer.
type RevocationProof struct {
	Issuer       common.Address
	Subject      common.Address
	RevokedBlock *big.Int
	Reason       [32]byte
}

func (rp *RevocationProof) String() string {
	return fmt.Sprintf("Issuer: %s, Subject: %s, RevokedBlock: %v, Reason: %v", rp.Issuer.Hex(), rp.Subject.Hex(), rp.RevokedBlock, rp.Reason)
}
