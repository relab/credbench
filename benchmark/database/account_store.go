package database

import (
	pb "github.com/relab/bbchain-dapp/benchmark/proto"
)

type AccountStore interface {
	Add(path []string, account *pb.Account) error
	AddAccounts(path []string, accounts []*pb.Account) error
	Get(path []string, key string) (*pb.Account, error)
	GetUnusedAccounts(path []string, n int) ([]*pb.Account, error)
	SelectAccounts(path []string, keys []string, selectType pb.Type) error
	GetAndSelect(path []string, n int, selectType pb.Type) ([]*pb.Account, error)
	All(path []string) ([]*pb.Account, error)
	DeleteAccount(path []string, key string) error
	DeleteBucket(path []string) error
}
