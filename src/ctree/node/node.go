package node

//go:generate abigen --abi ../../ethereum/build/abi/Node.abi --bin ../../ethereum/build/bin/Node.bin --pkg node --type NodeContract --out ./node_contract.go

//go:generate abigen --abi ../../ethereum/build/abi/NodeInterface.abi --bin ../../ethereum/build/bin/NodeInterface.bin --pkg node --type NodeInterface --out ./inode_contract.go

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/relab/ct-eth-dapp/src/ctree/owners"
	"github.com/relab/ct-eth-dapp/src/encode"
)

const (
	LeafRole uint8 = iota
	InnerRole
)

// Node is a Go wrapper around an node contract.
type Node struct {
	*owners.Owners
	contract *NodeContract
	address  common.Address
}

// NewNode creates a struct exposing convenient operations to
// interact with the Node contract.
func NewNode(contractAddr common.Address, backend bind.ContractBackend) (*Node, error) {
	n, err := NewNodeContract(contractAddr, backend)
	if err != nil {
		return nil, err
	}
	o, err := owners.NewOwners(contractAddr, backend)
	if err != nil {
		return nil, err
	}
	return &Node{o, n, contractAddr}, nil
}

// Address returns the contract address of the node.
func (n *Node) Address() common.Address {
	return n.address
}

// AddNode registers a new node contract
func (n *Node) AddNode(opts *bind.TransactOpts, node common.Address) (*types.Transaction, error) {
	return n.contract.AddChild(opts, node)
}

// GetRoot returns the aggregated proof of a subject
func (n *Node) GetRoot(opts *bind.CallOpts, subject common.Address) ([32]byte, error) {
	return n.contract.GetRoot(opts, subject)
}

// RegisterCredential issues a new credential proof ensuring append-only property.
// If the credential already exist, this function register the consentiment to // the given digest by the caller, if he is one of the contract's owners
func (n *Node) RegisterCredential(opts *bind.TransactOpts, subject common.Address, digest [32]byte, witnesses []common.Address) (*types.Transaction, error) {
	return n.contract.RegisterCredential(opts, subject, digest, witnesses)
}

// ApproveCredential approves the emission of a quorum signed credential proof
func (n *Node) ApproveCredential(opts *bind.TransactOpts, digest [32]byte) (*types.Transaction, error) {
	return n.contract.ApproveCredential(opts, digest)
}

// Revoke revokes an issued credential proof
func (n *Node) Revoke(opts *bind.TransactOpts, digest [32]byte, reason [32]byte) (*types.Transaction, error) {
	return n.contract.RevokeCredential(opts, digest, reason)
}

// AggregateCredentials aggregateCredentials aggregates the digests of a given subject.
func (n *Node) AggregateCredentials(opts *bind.TransactOpts, subject common.Address, digests [][32]byte) (*types.Transaction, error) {
	return n.contract.AggregateCredentials(opts, subject, digests)
}

// GetCredentialProof maps document digest to issued credential proof
func (n *Node) GetCredentialProof(opts *bind.CallOpts, digest [32]byte) *NotaryCredentialProof {
	proof, _ := n.contract.GetCredentialProof(opts, digest)
	var cp NotaryCredentialProof = proof
	return &cp
}

// GetRevokedProof maps document digest to revoked proof
func (n *Node) GetRevokedProof(opts *bind.CallOpts, digest [32]byte) *NotaryRevocationProof {
	proof, _ := n.contract.GetRevokedProof(opts, digest)
	var rp NotaryRevocationProof = proof
	return &rp
}

// IsSigned returns whether an owner already signed a digest
func (n *Node) IsSigned(opts *bind.CallOpts, digest [32]byte, owner common.Address) (bool, error) {
	return n.contract.IsSigned(opts, digest, owner)
}

// IsQuorumSigned verify if a credential proof was signed by a quorum
func (n *Node) IsQuorumSigned(opts *bind.CallOpts, digest [32]byte) (bool, error) {
	return n.contract.IsQuorumSigned(opts, digest)
}

// GetDigests returns the list of the issued credentials' digests of a subject
func (n *Node) GetDigests(opts *bind.CallOpts, subject common.Address) ([][32]byte, error) {
	return n.contract.GetDigests(opts, subject)
}

// GetWitnesses returns the witnesses of a proof
func (n *Node) GetWitnesses(opts *bind.CallOpts, digest [32]byte) ([]common.Address, error) {
	return n.contract.GetWitnesses(opts, digest)
}

// GetEvidenceRoot returns the root of the evidences of an issued credential proof
func (n *Node) GetEvidenceRoot(opts *bind.CallOpts, digest [32]byte) ([32]byte, error) {
	return n.contract.GetEvidenceRoot(opts, digest)
}

// GetRevoked returns a list of revoked credentials
func (n *Node) GetRevoked(opts *bind.CallOpts, subject common.Address) ([][32]byte, error) {
	return n.contract.GetRevoked(opts, subject)
}

// IsRevoked verifies if a credential proof was revoked
func (n *Node) IsRevoked(opts *bind.CallOpts, digest [32]byte) (bool, error) {
	return n.contract.IsRevoked(opts, digest)
}

// VerifyCredential checks whether the credential is valid
func (n *Node) VerifyCredential(onchain bool, opts *bind.CallOpts, subject common.Address, digest [32]byte) (bool, error) {
	if onchain {
		return n.contract.VerifyCredential(opts, subject, digest)
	}
	cp, err := n.contract.GetCredentialProof(opts, digest)
	if err != nil {
		return false, err
	}
	if cp.InsertedBlock == big.NewInt(0) {
		return false, fmt.Errorf("Credential %x not found", digest)
	}
	if cp.Subject != subject {
		return false, fmt.Errorf("Wrong subject %s", subject.Hex())
	}
	if !cp.Approved {
		return false, fmt.Errorf("Credential not approved")
	}
	signed, err := n.contract.IsQuorumSigned(opts, digest)
	if err != nil {
		return false, err
	}
	if !signed {
		q, err := n.contract.Quorum(opts)
		if err != nil {
			return false, err
		}
		return false, fmt.Errorf("Credential not signed by quorum. Require %d but has %d", q, cp.Signed)
	}
	revoked, err := n.contract.IsRevoked(opts, digest)
	if revoked {
		rp, err := n.contract.GetRevokedProof(opts, digest)
		if err != nil {
			return false, err
		}
		return false, fmt.Errorf("Credential revoked at %s", rp.RevokedBlock.String())
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// VerifyIssuedCredentials checks whether all credentials of a given subject are valid
func (n *Node) VerifyIssuedCredentials(onchain bool, opts *bind.CallOpts, subject common.Address) (bool, error) {
	if onchain {
		return n.contract.VerifyIssuedCredentials(opts, subject)
	}
	digests, err := n.contract.GetDigests(opts, subject)
	if err != nil {
		return false, err
	}
	for _, d := range digests {
		if ok, err := n.VerifyCredential(false, opts, subject, d); !ok || err != nil {
			return false, err
		}
	}
	return true, nil
}

// VerifyCredentialRoot
func (n *Node) VerifyCredentialRoot(onchain bool, opts *bind.CallOpts, subject common.Address, root [32]byte) (bool, error) {
	if onchain {
		return n.contract.VerifyCredentialRoot(opts, subject, root)
	}
	digests, err := n.contract.GetDigests(opts, subject)
	if err != nil {
		return false, err
	}
	if len(digests) == 0 {
		return false, nil
	}

	p, err := n.contract.GetProof(opts, subject)
	if err != nil {
		return false, err
	}
	if bytes.Equal(p.Proof[:], []byte{}) {
		return false, errors.New("root not found")
	}
	if p.Proof != root {
		return false, errors.New("wrong root given")
	}

	r, err := encode.EncodeByteArray(digests)
	if err != nil {
		return false, err
	}
	return bytes.Equal(r[:], root[:]), err
}

// TODO: check proof: InsertedBlock and BlockTimestamp

// VerifyCredentialTree performs a pre-order tree traversal over
// the credential tree of a given subject and verifies if the given
// root match with the current root on the root of the credential tree
// and if all the sub-trees were correctly built.
func (n *Node) VerifyCredentialTree(onchain bool, opts *bind.CallOpts, subject common.Address) (bool, error) {
	if onchain {
		return n.contract.VerifyCredentialTree(opts, subject)
	}
	return false, errors.New("not implemented")
}
