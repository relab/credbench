package datastore

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	proto "github.com/golang/protobuf/proto"
	"github.com/relab/ct-eth-dapp/benchmark/database"
	hu "github.com/relab/ct-eth-dapp/benchmark/hexutil"
	pb "github.com/relab/ct-eth-dapp/benchmark/proto"
)

var (
	ErrZeroAddress     = errors.New("zero address given")
	ErrNoAccountsFound = errors.New("no accounts found")
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

func GetAddresses(accounts []*pb.Account) []common.Address {
	addresses := make([]common.Address, len(accounts))
	for i, a := range accounts {
		addresses[i] = common.HexToAddress(a.Address)
	}
	return addresses
}

// PutAccount adds a new Account to the EthAccountStore
func (as *EthAccountStore) PutAccount(accounts ...*pb.Account) error {
	if len(accounts) < 1 {
		return ErrNoAccountsFound
	}

	for _, account := range accounts {
		value, err := proto.Marshal(account)
		if err != nil {
			return err
		}
		address := common.HexToAddress(account.Address)
		if address == (common.Address{}) {
			return ErrZeroAddress
		}
		err = as.ds.db.AddEntry(as.ds.sPath, address.Bytes(), value)
		if err != nil {
			return err
		}
	}
	return nil
}

// Get gets an account
func (as *EthAccountStore) GetAccount(key []byte) (*pb.Account, error) {
	account := &pb.Account{}
	buf, err := as.ds.db.GetEntry(as.ds.sPath, key)
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

	//TODO keep the last used key
	firstKey, firstValue, err := as.ds.db.GetFirstEntry(as.ds.sPath)
	if err != nil {
		return accounts, err
	}
	i := 0
	key, value := firstKey, firstValue
	for i < n {
		account := &pb.Account{}
		err = proto.Unmarshal(value, account)
		if err != nil {
			return accounts, err
		}

		if account.GetSelected() == pb.Type_NONE {
			accounts = append(accounts, account)
			i++
		}

		// Get next
		key, value, err = as.ds.db.GetNextEntry(as.ds.sPath, key)
		if err != nil || key == nil {
			return accounts, err
		}
	}
	if len(accounts) == 0 {
		return accounts, ErrNoAccountsFound
	}
	return accounts, err
}

func (as *EthAccountStore) SelectAccounts(keys [][]byte, selectType pb.Type) ([]*pb.Account, error) {
	var accounts []*pb.Account
	var err error
	for _, key := range keys {
		account, err := as.GetAccount(key)
		if err != nil {
			return accounts, err
		}

		account.Selected = selectType

		err = as.PutAccount(account)
		if err != nil {
			return accounts, err
		}

		accounts = append(accounts, account)
	}
	if len(accounts) == 0 {
		return accounts, ErrNoAccountsFound
	}
	return accounts, err
}

func (as EthAccountStore) GetAndSelect(n int, selectType pb.Type) ([]*pb.Account, error) {
	accounts, err := as.GetUnusedAccounts(n)
	if err != nil {
		return nil, err
	}
	accounts, err = as.SelectAccounts(hu.ByteAddresses(accounts), selectType)
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
	if len(accounts) == 0 {
		return accounts, ErrNoAccountsFound
	}
	return accounts, err
}

func (as *EthAccountStore) DeleteAccount(key []byte) error {
	return as.ds.db.DeleteEntry(as.ds.sPath, key)
}
