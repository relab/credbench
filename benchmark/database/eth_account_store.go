package database

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	bolt "go.etcd.io/bbolt"

	proto "github.com/golang/protobuf/proto"
	pb "github.com/relab/bbchain-dapp/benchmark/proto"
)

// Bucket("accounts")
// kv: eth_address -> Account (address, privkey, []contracts_address, []digest)
var ethAccountsBucket = "eth_accounts"

type EthAccountStore struct {
	DB   *Database
	Path []string
}

func CreateAccountStore(d *Database) (*EthAccountStore, error) {
	if err := d.CreateBucketPath([]string{ethAccountsBucket}); err != nil {
		return &EthAccountStore{}, err
	}
	return NewAccountStore(d, []string{ethAccountsBucket}), nil
}

func NewAccountStore(d *Database, path []string) *EthAccountStore {
	return &EthAccountStore{Path: path, DB: d}
}

// Add a new Account to the EthAccountStore
func (as *EthAccountStore) Add(path []string, account *pb.Account) error {
	if account == nil {
		return fmt.Errorf("account cannot be nil")
	}
	buf, err := proto.Marshal(account)
	if err != nil {
		return err
	}
	return as.DB.AddEntry(path, hexutil.MustDecode(account.Address), buf)
}

func (as *EthAccountStore) AddAccounts(path []string, accounts []*pb.Account) error {
	if len(accounts) < 1 {
		return fmt.Errorf("empty list of accounts given")
	}
	for _, a := range accounts {
		err := as.Add(path, a)
		if err != nil {
			return err
		}
	}
	return nil
}

// Get gets an account
func (as *EthAccountStore) Get(path []string, key string) (*pb.Account, error) {
	account := &pb.Account{}
	buf, err := as.DB.GetEntry(path, hexutil.MustDecode(key))
	if err != nil {
		return account, err
	}
	if buf != nil {
		err := proto.Unmarshal(buf, account)
		if err != nil {
			return account, err
		}
	}
	return account, err
}

func (as *EthAccountStore) GetUnusedAccounts(path []string, n int) ([]*pb.Account, error) {
	var accounts []*pb.Account
	err := as.DB.View(func(tx *bolt.Tx) error {
		b, err := GetBucket(tx, path)
		if err != nil {
			return err
		}
		c := b.Cursor()
		i := 0
		for k, v := c.First(); i < n && k != nil; k, v = c.Next() {
			if v != nil {
				account := &pb.Account{}
				err := proto.Unmarshal(v, account)
				if err != nil {
					return err
				}
				if account.GetSelected() == pb.Type_NONE {
					accounts = append(accounts, account)
					i++
				}
			}
		}
		return nil
	})
	return accounts, err
}

func (as *EthAccountStore) SelectAccounts(path []string, keys []string, selectType pb.Type) error {
	err := as.DB.Update(func(tx *bolt.Tx) error {
		b, err := GetBucket(tx, path)
		if err != nil {
			return err
		}
		for _, k := range keys {
			account := &pb.Account{}
			key := hexutil.MustDecode(k)
			buf := b.Get(key)
			if buf != nil {
				err := proto.Unmarshal(buf, account)
				if err != nil {
					return err
				}
			}
			account.Selected = selectType

			buf, err = proto.Marshal(account)
			if err != nil {
				return err
			}
			err = b.Put(key, buf)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (as EthAccountStore) GetAndSelect(path []string, n int, selectType pb.Type) ([]*pb.Account, error) {
	accounts, err := as.GetUnusedAccounts(path, n)
	if err != nil {
		return nil, err
	}
	err = as.SelectAccounts(path, HexAddresses(accounts), selectType)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

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

// All returns all accounts under the last bucket at path
func (as *EthAccountStore) All(path []string) ([]*pb.Account, error) {
	var accounts []*pb.Account
	err := as.DB.MapValues(path, func(value []byte) error {
		account := &pb.Account{}
		err := proto.Unmarshal(value, account)
		if err != nil {
			return err
		}
		accounts = append(accounts, account)
		return nil
	})
	return accounts, err
}

func (as *EthAccountStore) DeleteAccount(path []string, key string) error {
	return as.DB.DeleteEntry(path, hexutil.MustDecode(key))
}

func (as *EthAccountStore) DeleteBucket(path []string) error {
	return as.DB.DeleteBucket(path)
}
