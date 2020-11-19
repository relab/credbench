package ctree

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type Verifier interface {
	VerifyCredential(onchain bool, opts *bind.CallOpts, subject common.Address, digest [32]byte) (bool, error)

	VerifyIssuedCredentials(onchain bool, opts *bind.CallOpts, subject common.Address) (bool, error)

	VerifyCredentialRoot(onchain bool, opts *bind.CallOpts, subject common.Address, root [32]byte) (bool, error)

	VerifyCredentialTree(onchain bool, opts *bind.CallOpts, subject common.Address) (bool, error)
}
