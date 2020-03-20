package notary

//go:generate abigen --abi ../../ethereum/build/abi/Timed.abi --bin ../../ethereum/build/bin/Timed.bin --pkg contract --type Timed --out ./contract/timed.go

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/r0qs/bbchain-dapp/src/core/notary/contract"
)

var TimedParams = &Params{ContractAbi: contract.TimedABI}

// Timed is a Go wrapper around an on-chain timed contract.
type Timed struct {
	contract *contract.Timed
}

// NewTimed creates a struct exposing convenient operations to
// interact with the Timed contract.
func NewTimed(contractAddr common.Address, backend bind.ContractBackend) (*Timed, error) {
	t, err := contract.NewTimed(contractAddr, backend)
	if err != nil {
		return nil, err
	}
	return &Timed{contract: t}, nil
}

// StartingTime returns the contract starting time
func (t *Timed) StartingTime(opts *bind.CallOpts) (*big.Int, error) {
	return t.contract.StartingTime(opts)
}

// EndingTime returns the contract ending time
func (t *Timed) EndingTime(opts *bind.CallOpts) (*big.Int, error) {
	return t.contract.EndingTime(opts)
}

// IsStarted returna true if the contract is started, false otherwise
func (t *Timed) IsStarted(opts *bind.CallOpts) (bool, error) {
	return t.contract.IsStarted(opts)
}

// HasEnded checks whether the notarization period has already elapsed
func (t *Timed) HasEnded(opts *bind.CallOpts) (bool, error) {
	return t.contract.HasEnded(opts)
}

// StillRunning returns true if the notarization period not ended yet
func (t *Timed) StillRunning(opts *bind.CallOpts) (bool, error) {
	return t.contract.StillRunning(opts)
}
