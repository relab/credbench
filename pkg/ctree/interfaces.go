package ctree

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/relab/ct-eth-dapp/pkg/ctree/notary"
)

type Issuer interface {
	NodeInterface
	Verifier
}

type NodeInterface interface {
	Address() common.Address

	// RegisterCredential issues a new credential proof ensuring append-only property.
	RegisterCredential(opts *bind.TransactOpts, subject common.Address, digest [32]byte, witnesses []common.Address) (*types.Transaction, error)

	// Revoke revokes an issued credential proof
	Revoke(opts *bind.TransactOpts, digest [32]byte, reason [32]byte) (*types.Transaction, error)

	// ApproveCredential approves the emission of a quorum signed credential proof
	ApproveCredential(opts *bind.TransactOpts, digest [32]byte) (*types.Transaction, error)

	// AggregateCredentials aggregateCredentials aggregates the digests of a given subject.
	AggregateCredentials(opts *bind.TransactOpts, subject common.Address, digests [][32]byte) (*types.Transaction, error)

	// GetCredentialProof maps document digest to issued credential proof
	GetCredentialProof(opts *bind.CallOpts, digest [32]byte) *notary.NotaryCredentialProof

	// GetRevokedProof maps document digest to revoked proof
	GetRevokedProof(opts *bind.CallOpts, digest [32]byte) *notary.NotaryRevocationProof

	// GetRevoked returns a list of revoked credentials
	GetRevoked(opts *bind.CallOpts, subject common.Address) ([][32]byte, error)

	// IsRevoked verifies if a credential proof was revoked
	IsRevoked(opts *bind.CallOpts, digest [32]byte) (bool, error)

	// IsSigned returns whether an owner already signed a digest
	IsSigned(opts *bind.CallOpts, digest [32]byte, owner common.Address) (bool, error)

	// IsQuorumSigned verify if a credential proof was signed by a quorum
	IsQuorumSigned(opts *bind.CallOpts, digest [32]byte) (bool, error)

	// GetDigests returns the list of the issued credentials' digests of a subject
	GetDigests(opts *bind.CallOpts, subject common.Address) ([][32]byte, error)

	// GetWitnesses returns the witnesses of a proof
	GetWitnesses(opts *bind.CallOpts, digest [32]byte) ([]common.Address, error)

	// GetEvidenceRoot returns the root of the evidences of an issued credential proof
	GetEvidenceRoot(opts *bind.CallOpts, digest [32]byte) ([32]byte, error)

	// GetRoot returns the aggregated proof of a subject
	GetRoot(opts *bind.CallOpts, subject common.Address) ([32]byte, error)
}

type Verifier interface {
	VerifyCredential(onchain bool, opts *bind.CallOpts, subject common.Address, digest [32]byte) error

	VerifyIssuedCredentials(onchain bool, opts *bind.CallOpts, subject common.Address) error

	VerifyCredentialRoot(onchain bool, opts *bind.CallOpts, subject common.Address, root [32]byte) error

	VerifyCredentialTree(onchain bool, opts *bind.CallOpts, subject common.Address) error
}
