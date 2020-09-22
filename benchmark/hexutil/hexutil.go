package hexutil

import (
	"github.com/ethereum/go-ethereum/common"
	pb "github.com/relab/ct-eth-dapp/benchmark/proto"
)

func HexSliceToAddress(addresses []string) []common.Address {
	a := make([]common.Address, len(addresses))
	for i, addr := range addresses {
		a[i] = common.HexToAddress(addr)
	}
	return a
}

func HexAddresses(accounts []*pb.Account) []string {
	addresses := make([]string, len(accounts))
	for i, a := range accounts {
		addresses[i] = a.Address
	}
	return addresses
}

func ByteSliceToAddress(addresses [][]byte) []common.Address {
	a := make([]common.Address, len(addresses))
	for i, addr := range addresses {
		a[i] = common.BytesToAddress(addr)
	}
	return a
}

func ByteAddresses(accounts []*pb.Account) [][]byte {
	addresses := make([][]byte, len(accounts))
	for i, a := range accounts {
		addresses[i] = common.HexToAddress(a.Address).Bytes()
	}
	return addresses
}
