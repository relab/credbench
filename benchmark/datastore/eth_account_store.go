package datastore

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"

	proto "github.com/golang/protobuf/proto"
	"github.com/relab/ct-eth-dapp/benchmark/database"
	hu "github.com/relab/ct-eth-dapp/benchmark/hexutil"
	pb "github.com/relab/ct-eth-dapp/benchmark/proto"
)

// Bucket("accounts")
// kv: eth_address -> AccountProto (address, privkey, []contracts_address, []digest)
var ethAccountsBucket = "eth_accounts"

// EthAccountStore implements account store for ethereum accounts
type EthAccountStore struct {
	AccountStore
	ds DataStore
}

func CreateEthAccountStore(db database.Database) error {
	return db.CreateBucketPath([]string{ethAccountsBucket})
}

func NewEthAccountStore(db database.Database) *EthAccountStore {
	return &EthAccountStore{ds: DataStore{db: db, sPath: []string{ethAccountsBucket}}}
}

// Add a new Account to the EthAccountStore
func (as *EthAccountStore) Add(account *pb.Account) error {
	if account == nil {
		return fmt.Errorf("account cannot be nil")
	}
	buf, err := proto.Marshal(account)
	if err != nil {
		return err
	}
	return as.ds.db.AddEntry(as.ds.sPath, hexutil.MustDecode(account.Address), buf)
}

func (as *EthAccountStore) AddAccounts(accounts []*pb.Account) error {
	if len(accounts) < 1 {
		return fmt.Errorf("empty list of accounts given")
	}
	for _, a := range accounts {
		err := as.Add(a)
		if err != nil {
			return err
		}
	}
	return nil
}

// Get gets an account
func (as *EthAccountStore) Get(key string) (*pb.Account, error) {
	account := &pb.Account{}
	buf, err := as.ds.db.GetEntry(as.ds.sPath, hexutil.MustDecode(key))
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

func (as *EthAccountStore) GetUnusedAccounts(n int) ([]*pb.Account, error) {
	var accounts []*pb.Account

	i := 0
	err := as.ds.db.MapValues(as.ds.sPath, func(value []byte) error {
		if i == n {
			return nil // force break ?
		}
		account := &pb.Account{}
		err := proto.Unmarshal(value, account)
		if err != nil {
			return err
		}
		if account.GetSelected() == pb.Type_NONE {
			accounts = append(accounts, account)
			i++
		}
		return nil
	})
	return accounts, err
}

func (as *EthAccountStore) SelectAccounts(keys []string, selectType pb.Type) error {
	var err error
	for _, k := range keys {
		account := &pb.Account{}
		key := hexutil.MustDecode(k)

		err = as.ds.db.UpdateEntry(as.ds.sPath, key, func(value []byte) error {
			if value != nil {
				err := proto.Unmarshal(value, account)
				if err != nil {
					return err
				}
			}
			account.Selected = selectType

			value, err = proto.Marshal(account)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return err
}

func (as EthAccountStore) GetAndSelect(n int, selectType pb.Type) ([]*pb.Account, error) {
	accounts, err := as.GetUnusedAccounts(n)
	if err != nil {
		return nil, err
	}
	err = as.SelectAccounts(hu.HexAddresses(accounts), selectType)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

// All returns all accounts under the last bucket at path
func (as *EthAccountStore) All() ([]*pb.Account, error) {
	var accounts []*pb.Account
	err := as.ds.db.MapValues(as.ds.sPath, func(value []byte) error {
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

func (as *EthAccountStore) DeleteAccount(key string) error {
	return as.ds.db.DeleteEntry(as.ds.sPath, hexutil.MustDecode(key))
}
