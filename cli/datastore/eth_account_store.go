package datastore

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	proto "github.com/golang/protobuf/proto"
	"github.com/relab/ct-eth-dapp/cli/database"
	pb "github.com/relab/ct-eth-dapp/cli/proto"
)

var (
	errZeroAddress     = errors.New("zero address given")
	errNoAccountsFound = errors.New("no accounts found")
)

// Bucket("accounts")
// kv: eth_address -> AccountProto (address, privkey, []contracts_address, []digest)
var ethAccountsBucket = "eth_accounts"

// EthAccountStore implements account store for ethereum accounts
type EthAccountStore struct {
	ds DataStore
}

func CreateEthAccountStore(db *database.BoltDB) error {
	return db.CreateBucketPath(ethAccountsBucket)
}

func NewEthAccountStore(db *database.BoltDB) *EthAccountStore {
	return &EthAccountStore{ds: DataStore{db: db, path: ethAccountsBucket}}
}

// PutAccount adds a new Account to the EthAccountStore
func (as *EthAccountStore) PutAccount(accounts ...*pb.Account) error {
	if len(accounts) < 1 {
		return errNoAccountsFound
	}

	for _, account := range accounts {
		value, err := proto.Marshal(account)
		if err != nil {
			return err
		}
		address := common.HexToAddress(account.HexAddress)
		if address == (common.Address{}) {
			return errZeroAddress
		}
		err = as.ds.db.AddEntry(as.ds.path, address.Bytes(), value)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetAccount gets an account
func (as EthAccountStore) GetAccount(key []byte) (*pb.Account, error) {
	account := &pb.Account{}
	buf, err := as.ds.db.GetEntry(as.ds.path, key)
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

// GetUnusedAccounts return `n` free accounts
func (as EthAccountStore) GetUnusedAccounts(n int) (Accounts, error) {
	var accounts Accounts

	//TODO keep the last used key
	firstKey, firstValue, err := as.ds.db.GetFirstEntry(as.ds.path)
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
		key, value, err = as.ds.db.GetNextEntry(as.ds.path, key)
		if err != nil || key == nil {
			return accounts, err
		}
	}
	if len(accounts) == 0 || len(accounts) < n {
		return accounts, errNoAccountsFound
	}
	return accounts, err
}

// SelectAccount selects n accounts of the same type
func (as *EthAccountStore) SelectAccount(selectType pb.Type, keys ...[]byte) (Accounts, error) {
	var accounts Accounts
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
	if len(accounts) != len(keys) {
		return accounts, errNoAccountsFound
	}
	return accounts, err
}

func (as EthAccountStore) GetAndSelect(n int, selectType pb.Type) (Accounts, error) {
	accounts, err := as.GetUnusedAccounts(n)
	if err != nil {
		return nil, err
	}
	accounts, err = as.SelectAccount(selectType, accounts.ToBytes()...)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

func (as EthAccountStore) GetByType(n int, selectType pb.Type) (Accounts, error) {
	var accounts Accounts
	firstKey, firstValue, err := as.ds.db.GetFirstEntry(as.ds.path)
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

		if account.GetSelected() == selectType {
			accounts = append(accounts, account)
			i++
		}

		// Get next
		key, value, err = as.ds.db.GetNextEntry(as.ds.path, key)
		if err != nil || key == nil {
			return accounts, err
		}
	}
	if len(accounts) == 0 || len(accounts) < n {
		return accounts, errNoAccountsFound
	}
	return accounts, err
}

// All returns all accounts under the last bucket at path
func (as EthAccountStore) All() (Accounts, error) {
	var accounts Accounts
	err := as.ds.db.MapValues(as.ds.path, func(value []byte) error {
		account := &pb.Account{}
		err := proto.Unmarshal(value, account)
		if err != nil {
			return err
		}
		accounts = append(accounts, account)
		return nil
	})
	if len(accounts) == 0 {
		return accounts, errNoAccountsFound
	}
	return accounts, err
}

func (as *EthAccountStore) DeleteAccount(keys ...[]byte) error {
	for _, key := range keys {
		err := as.ds.db.DeleteEntry(as.ds.path, key)
		if err != nil {
			return err
		}
	}
	return nil
}
