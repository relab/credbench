package datastore

import (
	pb "github.com/relab/ct-eth-dapp/benchmark/proto"
)

type AccountStore interface {
	Add(account *pb.Account) error
	AddAccounts(accounts []*pb.Account) error

	Get(key string) (*pb.Account, error)
	GetUnusedAccounts(n int) ([]*pb.Account, error)
	GetAndSelect(n int, selectType pb.Type) ([]*pb.Account, error)

	All() ([]*pb.Account, error)
	SelectAccounts(keys []string, selectType pb.Type) error

	DeleteAccount(key string) error
}
