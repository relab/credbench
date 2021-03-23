package node

import (
	"bytes"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/relab/ct-eth-dapp/pkg/ctree"
	"github.com/relab/ct-eth-dapp/pkg/ctree/notary"
	"github.com/relab/ct-eth-dapp/pkg/ctree/owners"
	"github.com/relab/ct-eth-dapp/pkg/deployer"
	"github.com/relab/ct-eth-dapp/pkg/encode"

	bindings "github.com/relab/bbchain-bindings/node"
)

// FIXME: duplicated in ctree
const (
	LeafRole uint8 = iota
	InnerRole
)

var (
	ErrVerificationFailed    = errors.New("verification failed")
	ErrNoCredentials         = errors.New("there is no credentials")
	ErrCredentialNotFound    = errors.New("credential not found")
	ErrWrongSubject          = errors.New("wrong subject")
	ErrCredentialNotApproved = errors.New("credential not approved")
	ErrNotQuorumSigned       = errors.New("credential not signed by quorum")
	ErrCredentialRevoked     = errors.New("credential revoked")
	ErrWrongRoot             = errors.New("root does not match")
	ErrRootNotFound          = errors.New("root not found")
)

// Node is a Go wrapper around an node contract.
type Node struct {
	*owners.Owners
	contract *bindings.Node
	address  common.Address
	backend  bind.ContractBackend
}

// NewNode creates a struct exposing convenient operations to
// interact with the Node contract.
func NewNode(contractAddr common.Address, backend bind.ContractBackend) (*Node, error) {
	n, err := bindings.NewNode(contractAddr, backend)
	if err != nil {
		return nil, err
	}
	o, err := owners.NewOwners(contractAddr, backend)
	if err != nil {
		return nil, err
	}
	return &Node{o, n, contractAddr, backend}, nil
}

func Deploy(auth *bind.TransactOpts, backend bind.ContractBackend, libs map[string]string, params ...interface{}) (common.Address, *types.Transaction, *Node, error) {
	contractBin := deployer.LinkContract(bindings.NodeBin, libs)

	address, tx, _, err := deployer.DeployContract(auth, backend, bindings.NodeABI, contractBin, params)
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	n, err := NewNode(address, backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, n, nil
}

// Address returns the contract address of the node.
func (n *Node) Address() common.Address {
	return n.address
}

// AddNode registers a new node contract
func (n *Node) AddNode(opts *bind.TransactOpts, node common.Address) (*types.Transaction, error) {
	return n.contract.AddChild(opts, node)
}

func (n *Node) IsLeaf(opts *bind.CallOpts) (bool, error) {
	return n.contract.IsLeaf(opts)
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
func (n *Node) GetCredentialProof(opts *bind.CallOpts, digest [32]byte) *notary.NotaryCredentialProof {
	proof, _ := n.contract.GetCredentialProof(opts, digest)
	cp := notary.NotaryCredentialProof(proof)
	return &cp
}

// GetRevokedProof maps document digest to revoked proof
func (n *Node) GetRevokedProof(opts *bind.CallOpts, digest [32]byte) *notary.NotaryRevocationProof {
	proof, _ := n.contract.GetRevokedProof(opts, digest)
	rp := notary.NotaryRevocationProof(proof)
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
func (n *Node) VerifyCredential(onchain bool, opts *bind.CallOpts, subject common.Address, digest [32]byte) error {
	if onchain {
		ok, err := n.contract.VerifyCredential(opts, subject, digest)
		if err != nil {
			return err
		}
		if !ok {
			return ErrVerificationFailed
		}
		return nil
	}
	cp, err := n.contract.GetCredentialProof(opts, digest)
	if err != nil {
		return err
	}
	if cp.InsertedBlock == big.NewInt(0) {
		return ErrCredentialNotFound
	}
	if cp.Subject != subject {
		return ErrWrongSubject
	}
	if !cp.Approved {
		return ErrCredentialNotApproved
	}
	signed, err := n.contract.IsQuorumSigned(opts, digest)
	if err != nil {
		return err
	}
	if !signed {
		return ErrNotQuorumSigned
	}
	revoked, err := n.contract.IsRevoked(opts, digest)
	if err != nil {
		return err
	}
	if revoked {
		return ErrCredentialRevoked
	}
	return nil
}

// VerifyIssuedCredentials checks whether all credentials of a given subject are valid
func (n *Node) VerifyIssuedCredentials(onchain bool, opts *bind.CallOpts, subject common.Address) error {
	if onchain {
		ok, err := n.contract.VerifyIssuedCredentials(opts, subject)
		if err != nil {
			return err
		}
		if !ok {
			return ErrVerificationFailed
		}
		return nil
	}
	digests, err := n.contract.GetDigests(opts, subject)
	if err != nil {
		return err
	}
	if len(digests) == 0 {
		return ErrNoCredentials
	}
	for _, d := range digests {
		err := n.VerifyCredential(false, opts, subject, d)
		if err != nil {
			return err
		}
	}
	return nil
}

// VerifyCredentialRoot
func (n *Node) VerifyCredentialRoot(onchain bool, opts *bind.CallOpts, subject common.Address, root [32]byte) error {
	if onchain {
		ok, err := n.contract.VerifyCredentialRoot(opts, subject, root)
		if err != nil {
			return err
		}
		if !ok {
			return ErrVerificationFailed
		}
		return nil
	}
	digests, err := n.contract.GetDigests(opts, subject)
	if err != nil {
		return err
	}
	if len(digests) == 0 {
		return ErrNoCredentials
	}

	p, err := n.contract.GetProof(opts, subject)
	if err != nil {
		return err
	}
	if bytes.Equal(p.Proof[:], []byte{}) {
		return ErrRootNotFound
	}
	if p.Proof != root {
		return ErrWrongRoot
	}

	r, err := encode.EncodeByteArray(digests)
	if err != nil {
		return err
	}
	if bytes.Equal(r[:], root[:]) {
		return nil
	}
	return ErrVerificationFailed
}

// TODO: check proof: InsertedBlock and BlockTimestamp

// VerifyCredentialTree performs a pre-order tree traversal over
// the credential tree of a given subject and verifies if the given
// root match with the current root on the root of the credential tree
// and if all the sub-trees were correctly built.
func verifyCredentialTree(n ctree.Issuer, backend bind.ContractBackend, opts *bind.CallOpts, subject common.Address) error {
	digests, err := n.GetDigests(opts, subject)
	if err != nil {
		return err
	}
	if len(digests) == 0 {
		return ErrNoCredentials
	}
	for _, d := range digests {
		err := n.VerifyCredential(false, opts, subject, d)
		if err != nil {
			return err
		}
		c := n.GetCredentialProof(opts, d)
		for _, w := range c.Witnesses {
			node, err := NewNode(w, backend) // witnesses must be nodes
			if err != nil {
				return err
			}
			ok, err := node.IsLeaf(opts)
			if ok && err == nil {
				r, err := node.GetRoot(opts, subject)
				if err != nil {
					return err
				}
				err = node.VerifyCredentialRoot(false, opts, subject, r)
				if err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else {
				// check sub-tree
				err := verifyCredentialTree(node, backend, opts, subject)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (n *Node) VerifyCredentialTree(onchain bool, opts *bind.CallOpts, subject common.Address) error {
	if onchain {
		ok, err := n.contract.VerifyCredentialTree(opts, subject)
		if err != nil {
			return err
		}
		if !ok {
			return ErrVerificationFailed
		}
		return nil
	}
	return verifyCredentialTree(n, n.backend, opts, subject)
}

// TODO: add methods to verify json credential
// given a credential, parse the json and query the contracts cross-checking the data.
