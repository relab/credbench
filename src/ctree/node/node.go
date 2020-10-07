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

// RegisterCredential collects all subject's credentials and issue a
// new credential proof iff the aggregation of those credentials on
// the sub-contracts match the given root (i.e. off-chain aggregation == on-chain aggregation)
func (n *Node) RegisterCredential0(opts *bind.TransactOpts, subject common.Address, digest [32]byte, witnesses []common.Address) (*types.Transaction, error) {
	return n.contract.RegisterCredential(opts, subject, digest, witnesses)
}

// VerifyCredentialRoot
func (n *Node) VerifyCredentialRoot(opts *bind.CallOpts, subject common.Address, root [32]byte) (bool, error) {
	return n.contract.VerifyCredentialRoot(opts, subject, root)
}

// VerifyCredentialTree performs a pre-order tree traversal over
// the credential tree of a given subject and verifies if the given
// root match with the current root on the root of the credential tree
// and if all the sub-trees were correctly built.
func (n *Node) VerifyCredentialTree(opts *bind.CallOpts, subject common.Address) (bool, error) {
	return n.contract.VerifyCredentialTree(opts, subject)
}
