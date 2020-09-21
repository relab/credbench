package datastore

import (
	"errors"

	pb "github.com/relab/ct-eth-dapp/benchmark/proto"
)

var ErrEmptyData = errors.New("atempt to store empty data")

type AccountStore interface {
	PutAccount(account ...*pb.Account) error
	GetAccount(key []byte) (*pb.Account, error)
	GetUnusedAccounts(n int) ([]*pb.Account, error)
	GetAndSelect(n int, selectType pb.Type) ([]*pb.Account, error)
	All() ([]*pb.Account, error)
	SelectAccounts(keys [][]byte, selectType pb.Type) ([]*pb.Account, error)
	DeleteAccount(key []byte) error
}
