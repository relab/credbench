package notary

//go:generate abigen --abi ../../ethereum/build/abi/Issuer.abi --bin ../../ethereum/build/bin/Issuer.bin --pkg contract --type Issuer --out ./contract/issuer.go

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/relab/bbchain-dapp/src/core/notary/contract"
)

var IssuerParams = &Params{ContractABI: contract.IssuerABI}

// Issuer is a Go wrapper around an on-chain issuer contract.
type Issuer struct {
	*Owners
	address  common.Address
	contract *contract.Issuer
}

// NewIssuer creates a struct exposing convenient operations to
// interact with the Issuer contract.
func NewIssuer(contractAddr common.Address, backend bind.ContractBackend) (*Issuer, error) {
	i, err := contract.NewIssuer(contractAddr, backend)
	if err != nil {
		return nil, err
	}
	o, err := NewOwners(contractAddr, backend)
	return &Issuer{o, contractAddr, i}, nil
}

// Address returns the contract address of the issuer.
func (i *Issuer) Address() common.Address {
	return i.address
}

// Nonce is an incremental-only counter for issued credentials per subject
func (i *Issuer) Nonce(opts *bind.CallOpts, subject common.Address) (*big.Int, error) {
	return i.contract.Nonce(opts, subject)
}

// IssuedCredentials maps document digest to issued credential proof
func (i *Issuer) IssuedCredentials(opts *bind.CallOpts, digest [32]byte) *CredentialProof {
	proof, _ := i.contract.IssuedCredentials(opts, digest)
	var cp CredentialProof = proof
	return &cp
}

// RevokedCredentials maps document digest to revoked proof
func (i *Issuer) RevokedCredentials(opts *bind.CallOpts, digest [32]byte) *RevocationProof {
	proof, _ := i.contract.RevokedCredentials(opts, digest)
	var rp RevocationProof = proof
	return &rp
}

// OwnersSigned maps digest to owners that already signed it
func (i *Issuer) OwnersSigned(opts *bind.CallOpts, digest [32]byte, owner common.Address) (bool, error) {
	return i.contract.OwnersSigned(opts, digest, owner)
}

// DigestsBySubject returns the registered digests of a subject
func (i *Issuer) DigestsBySubject(opts *bind.CallOpts, subject common.Address) ([][32]byte, error) {
	return i.contract.DigestsBySubject(opts, subject)
}

// GetProof returns the aggregated proof of a subject
func (i *Issuer) GetProof(opts *bind.CallOpts, subject common.Address) ([32]byte, error) {
	return i.contract.GetProof(opts, subject)
}

// IsRevoked verifies if a credential proof was revoked
func (i *Issuer) IsRevoked(opts *bind.CallOpts, digest [32]byte) (bool, error) {
	return i.contract.IsRevoked(opts, digest)
}

// Certified verifies if a digest was already signed by all parties
func (i *Issuer) Certified(opts *bind.CallOpts, digest [32]byte) (bool, error) {
	return i.contract.Certified(opts, digest)
}

// RegisterCredential issues a new credential proof ensuring append-only property
func (i *Issuer) RegisterCredential(opts *bind.TransactOpts, subject common.Address, digest [32]byte) (*types.Transaction, error) {
	return i.contract.RegisterCredential(opts, subject, digest)
}

// ConfirmCredential confirms the emission of a quorum signed credential proof
func (i *Issuer) ConfirmCredential(opts *bind.TransactOpts, digest [32]byte) (*types.Transaction, error) {
	return i.contract.ConfirmCredential(opts, digest)
}

// Revoke revokes an issued credential proof
func (i *Issuer) Revoke(opts *bind.TransactOpts, digest [32]byte, reason [32]byte) (*types.Transaction, error) {
	return i.contract.RevokeCredential(opts, digest, reason)
}

// CheckCredentials verifies if a list of digests are certified
func (i *Issuer) CheckCredentials(opts *bind.CallOpts, digests [][32]byte) (bool, error) {
	return i.contract.CheckCredentials(opts, digests)
}

// AggregateCredentials aggregates the digests of a given subject on the credential level
func (i *Issuer) AggregateCredentials(opts *bind.TransactOpts, subject common.Address) (*types.Transaction, error) {
	return i.contract.AggregateCredentials(opts, subject)
}

// VerifyCredential verifies if the credential of a given subject was correctly generated
func (i *Issuer) VerifyCredential(opts *bind.CallOpts, subject common.Address, digest [32]byte) (bool, error) {
	return i.contract.VerifyCredential(opts, subject, digest)
}
