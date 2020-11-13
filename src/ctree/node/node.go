package node

//go:generate abigen --abi ../../ethereum/build/abi/Node.abi --bin ../../ethereum/build/bin/Node.bin --pkg node --type NodeContract --out ./node_contract.go

//go:generate abigen --abi ../../ethereum/build/abi/NodeInterface.abi --bin ../../ethereum/build/bin/NodeInterface.bin --pkg node --type NodeInterface --out ./inode_contract.go

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const (
	LeafRole uint8 = iota
	InnerRole
)

// Node is a Go wrapper around an node contract.
type Node struct {
	address  common.Address
	contract *NodeContract
}

// NewNode creates a struct exposing convenient operations to
// interact with the Node contract.
func NewNode(contractAddr common.Address, backend bind.ContractBackend) (*Node, error) {
	n, err := NewNodeContract(contractAddr, backend)
	if err != nil {
		return nil, err
	}
	return &Node{contractAddr, n}, nil
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

// OnVerifyCredentialRoot
func (n *Node) OnVerifyCredentialRoot(opts *bind.CallOpts, subject common.Address, root [32]byte) (bool, error) {
	return n.contract.VerifyCredentialRoot(opts, subject, root)
}

func (n *Node) OffVerifyCredentialRoot(opts *bind.CallOpts, subject common.Address, root [32]byte) (bool, error) {
	// TODO: implement me
	return false, nil
}

// VerifyCredentialTree performs a pre-order tree traversal over
// the credential tree of a given subject and verifies if the given
// root match with the current root on the root of the credential tree
// and if all the sub-trees were correctly built.
func (n *Node) OnVerifyCredentialTree(opts *bind.CallOpts, subject common.Address) (bool, error) {
	return n.contract.VerifyCredentialTree(opts, subject)
}

func (n *Node) OffVerifyCredentialTree(opts *bind.CallOpts, subject common.Address) (bool, error) {
	// TODO: implement me
	return false, nil
}
