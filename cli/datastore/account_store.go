package datastore

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	pb "github.com/relab/ct-eth-dapp/cli/proto"
)

var ErrEmptyData = errors.New("attempt to store empty data")

// Accounts is a list of Accounts
type Accounts []*pb.Account

func (accounts Accounts) ToBytes() [][]byte {
	addresses := make([][]byte, len(accounts))
	for i, a := range accounts {
		addresses[i] = common.HexToAddress(a.HexAddress).Bytes()
	}
	return addresses
}

func (accounts Accounts) ToHex() []string {
	addresses := make([]string, len(accounts))
	for i, a := range accounts {
		addresses[i] = a.HexAddress
	}
	return addresses
}

func (accounts Accounts) ToETHAddress() []common.Address {
	addresses := make([]common.Address, len(accounts))
	for i, a := range accounts {
		addresses[i] = common.HexToAddress(a.HexAddress)
	}
	return addresses
}
