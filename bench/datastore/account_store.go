package datastore

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	pb "github.com/relab/ct-eth-dapp/bench/proto"
)

var ErrEmptyData = errors.New("attempt to store empty data")

func ToETHAddress(byteAddresses [][]byte) []common.Address {
	addresses := make([]common.Address, len(byteAddresses))
	for i, a := range byteAddresses {
		addresses[i] = common.BytesToAddress(a)
	}
	return addresses
}

func ToHex(byteAddresses [][]byte) []string {
	addresses := make([]string, len(byteAddresses))
	for i, a := range byteAddresses {
		addresses[i] = hexutil.Encode(a)
	}
	return addresses
}

func AddressToBytes(address []common.Address) [][]byte {
	addresses := make([][]byte, len(address))
	for i, a := range address {
		addresses[i] = a.Bytes()
	}
	return addresses
}

// Accounts is a list of Accounts
type Accounts []*pb.Account

// ToBytes returns the accounts as a list of addresses in bytes
func (accounts Accounts) ToBytes() [][]byte {
	addresses := make([][]byte, len(accounts))
	for i, a := range accounts {
		addresses[i] = a.Address
	}
	return addresses
}

func (accounts Accounts) ToHex() []string {
	return ToHex(accounts.ToBytes())
}

func (accounts Accounts) ToETHAddress() []common.Address {
	return ToETHAddress(accounts.ToBytes())
}
