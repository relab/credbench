package issuer

//go:generate abigen --abi ../../../ethereum/build/abi/Issuer.abi --bin ../../../ethereum/build/bin/Issuer.bin --pkg issuer --type IssuerContract --out ./issuer_contract.go

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/relab/ct-eth-dapp/src/ctree/owners"
)

// Issuer is a Go wrapper around an issuer contract.
type Issuer struct {
	*owners.Owners
	address  common.Address
	contract *IssuerContract
}

// NewIssuer creates a struct exposing convenient operations to
// interact with the Issuer contract.
func NewIssuer(contractAddr common.Address, backend bind.ContractBackend) (*Issuer, error) {
	i, err := NewIssuerContract(contractAddr, backend)
	if err != nil {
		return nil, err
	}
	o, err := owners.NewOwners(contractAddr, backend)
	return &Issuer{o, contractAddr, i}, nil
}

// Address returns the contract address of the issuer.
func (i *Issuer) Address() common.Address {
	return i.address
}

// GetCredentialProof maps document digest to issued credential proof
func (i *Issuer) GetCredentialProof(opts *bind.CallOpts, digest [32]byte) *NotaryCredentialProof {
	proof, _ := i.contract.GetCredentialProof(opts, digest)
	var cp NotaryCredentialProof = proof
	return &cp
}

// GetRevokedProof maps document digest to revoked proof
func (i *Issuer) GetRevokedProof(opts *bind.CallOpts, digest [32]byte) *NotaryRevocationProof {
	proof, _ := i.contract.GetRevokedProof(opts, digest)
	var rp NotaryRevocationProof = proof
	return &rp
}

// // OwnersList return the list of owners
// func (i *Issuer) OwnersList(opts *bind.CallOpts) ([]common.Address, error) {
// 	return i.contract.Owners(opts)
// }

// IsSigned returns whether an owner already signed a digest
func (i *Issuer) IsSigned(opts *bind.CallOpts, digest [32]byte, owner common.Address) (bool, error) {
	return i.contract.IsSigned(opts, digest, owner)
}

// IsQuorumSigned verify if a credential proof was signed by a quorum
func (i *Issuer) IsQuorumSigned(opts *bind.CallOpts, digest [32]byte) (bool, error) {
	return i.contract.IsQuorumSigned(opts, digest)
}

// GetDigests returns the list of the issued credentials' digests of a subject
func (i *Issuer) GetDigests(opts *bind.CallOpts, subject common.Address) ([][32]byte, error) {
	return i.contract.GetDigests(opts, subject)
}

// GetRootProof returns the aggregated proof of a subject
func (i *Issuer) GetRootProof(opts *bind.CallOpts, subject common.Address) ([32]byte, error) {
	return i.contract.GetRootProof(opts, subject)
}

// GetWitnesses returns the witnesses of a proof
func (i *Issuer) GetWitnesses(opts *bind.CallOpts, digest [32]byte) ([]common.Address, error) {
	return i.contract.GetWitnesses(opts, digest)
}

// GetEvidenceRoot returns the root of the evidences of an issued credential proof
func (i *Issuer) GetEvidenceRoot(opts *bind.CallOpts, digest [32]byte) ([32]byte, error) {
	return i.contract.GetEvidenceRoot(opts, digest)
}

// GetRevoked returns a list of revoked credentials
func (i *Issuer) GetRevoked(opts *bind.CallOpts, subject common.Address) ([]common.Address, error) {
	return i.contract.GetRevoked(opts, subject)
}

// IsRevoked verifies if a credential proof was revoked
func (i *Issuer) IsRevoked(opts *bind.CallOpts, digest [32]byte) (bool, error) {
	return i.contract.IsRevoked(opts, digest)
}

// RegisterCredential issues a new credential proof ensuring append-only property
func (i *Issuer) RegisterCredential(opts *bind.TransactOpts, subject common.Address, digest [32]byte, root [32]byte, witnesses []common.Address) (*types.Transaction, error) {
	return i.contract.RegisterCredential(opts, subject, digest, root, witnesses)
}

// ConfirmCredential confirms the emission of a quorum signed credential proof
func (i *Issuer) ConfirmCredential(opts *bind.TransactOpts, digest [32]byte) (*types.Transaction, error) {
	return i.contract.ConfirmCredential(opts, digest)
}

// Revoke revokes an issued credential proof
func (i *Issuer) Revoke(opts *bind.TransactOpts, digest [32]byte, reason [32]byte) (*types.Transaction, error) {
	return i.contract.RevokeCredential(opts, digest, reason)
}

// VerifyCredential checks whether the credential is valid.
func (i *Issuer) VerifyCredential(opts *bind.CallOpts, subject common.Address, digest [32]byte) (bool, error) {
	return i.contract.VerifyCredential(opts, subject, digest)
}

// VerifyIssuedCredentials checks whether all credentials of a given subject are valid.
func (i *Issuer) VerifyIssuedCredentials(opts *bind.CallOpts, subject common.Address) (bool, error) {
	return i.contract.VerifyIssuedCredentials(opts, subject)
}

// AggregateCredentials aggregateCredentials aggregates the digests of a given subject.
func (i *Issuer) AggregateCredentials(opts *bind.TransactOpts, subject common.Address, digests [][32]byte) (*types.Transaction, error) {
	return i.contract.AggregateCredentials(opts, subject, digests)
}

// VerifyCredentialRoot checks whether the root exists and was correctly built based on the existent tree.
func (i *Issuer) VerifyCredentialRoot(opts *bind.CallOpts, subject common.Address, root [32]byte) (bool, error) {
	return i.contract.VerifyCredentialRoot(opts, subject, root)
}
