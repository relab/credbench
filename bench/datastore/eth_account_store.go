package datastore

import (
	"context"
	"errors"
	"math/big"
	"sync"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/relab/credbench/bench/database"

	pb "github.com/relab/credbench/bench/proto"
	ctaccounts "github.com/relab/credbench/pkg/accounts"
	proto "google.golang.org/protobuf/proto"
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
	lock    sync.Mutex
	ds      DataStore
	chainID *big.Int
}

func CreateEthAccountStore(db *database.BoltDB) error {
	return db.CreateBucketPath(ethAccountsBucket)
}

func NewEthAccountStore(db *database.BoltDB, chainID *big.Int) *EthAccountStore {
	return &EthAccountStore{
		ds: DataStore{
			db:   db,
			path: ethAccountsBucket,
		},
		chainID: chainID,
	}
}

func (as *EthAccountStore) GetTxOpts(key []byte, backend bind.ContractBackend) (*bind.TransactOpts, error) {
	as.lock.Lock()
	defer as.lock.Unlock()

	gasPrice, err := backend.SuggestGasPrice(context.TODO())
	if err != nil {
		return nil, err
	}

	account, err := as.GetAccount(key)
	if err != nil {
		return nil, err
	}

	address := common.BytesToAddress(account.Address)
	nonce, err := backend.PendingNonceAt(context.TODO(), address)
	if err != nil {
		return nil, err
	}
	if account.Nonce < nonce {
		account.Nonce = nonce
	}

	pk := ctaccounts.HexToKey(account.HexKey)
	transactOpts, err := bind.NewKeyedTransactorWithChainID(pk, as.chainID)
	if err != nil {
		return nil, err
	}
	transactOpts.GasLimit = uint64(6721975) // FIXME: get from config file
	transactOpts.GasPrice = gasPrice
	transactOpts.Nonce = new(big.Int).SetUint64(account.Nonce)

	err = as.incNonce(account)
	if err != nil {
		return nil, err
	}

	return transactOpts, nil
}

func (as *EthAccountStore) incNonce(account *pb.Account) error {
	atomic.AddUint64(&account.Nonce, 1)
	err := as.PutAccount(account)
	if err != nil {
		return err
	}
	return nil
}

// PutAccount adds a new Account to the EthAccountStore
func (as *EthAccountStore) PutAccount(accounts ...*pb.Account) error {
	if len(accounts) < 1 {
		return ErrNoAccountsFound
	}

	for _, account := range accounts {
		address := common.BytesToAddress(account.Address)
		if address == (common.Address{}) {
			return ErrZeroAddress
		}
		value, err := proto.Marshal(account)
		if err != nil {
			return err
		}
		err = as.ds.db.Put(as.ds.path, address.Bytes(), value)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetAccount gets an account
func (as *EthAccountStore) GetAccount(key []byte) (*pb.Account, error) {
	account := &pb.Account{}
	buf, err := as.ds.db.Get(as.ds.path, key)
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

func (as *EthAccountStore) GetUnusedAccounts(n int) (Accounts, error) {
	as.lock.Lock()
	defer as.lock.Unlock()

	var accounts Accounts
	if err := as.ds.db.Iterate(as.ds.path, n, func(value []byte) (bool, error) {
		account := &pb.Account{}
		err := proto.Unmarshal(value, account)
		if err != nil {
			return false, err
		}

		if account.GetSelected() == pb.Type_NONE {
			accounts = append(accounts, account)
			return true, nil
		}
		return false, nil
	}); err != nil {
		return accounts, err
	}
	if len(accounts) == 0 || len(accounts) < n {
		return accounts, ErrNoAccountsFound
	}
	return accounts, nil
}

func (as *EthAccountStore) SelectAccount(selectType pb.Type, keys ...[]byte) (Accounts, error) {
	as.lock.Lock()
	defer as.lock.Unlock()

	var accounts Accounts
	var err error
	for _, key := range keys {
		account := &pb.Account{}
		err = as.ds.db.Update(as.ds.path, key, func(value []byte) ([]byte, error) {
			if value == nil {
				return nil, nil // value does not exists or is bucket
			}
			err := proto.Unmarshal(value, account)
			if err != nil {
				return nil, err
			}
			account.Selected = selectType

			value, err = proto.Marshal(account)
			if err != nil {
				return nil, err
			}
			return value, nil
		})
		if err == nil && account.Selected != pb.Type_NONE {
			accounts = append(accounts, account)
		}
	}
	if len(accounts) != len(keys) {
		return accounts, ErrNoAccountsFound
	}
	return accounts, err
}

func (as *EthAccountStore) GetAccounts(keys ...[]byte) (Accounts, error) {
	as.lock.Lock()
	defer as.lock.Unlock()

	var accounts Accounts
	var err error
	for _, key := range keys {
		account, err := as.GetAccount(key)
		if err == nil {
			accounts = append(accounts, account)
		}
	}
	if len(accounts) != len(keys) {
		return accounts, ErrNoAccountsFound
	}
	return accounts, err
}

func (as *EthAccountStore) GetAllKeys(selectType pb.Type) ([][]byte, error) {
	as.lock.Lock()
	defer as.lock.Unlock()

	keys, err := as.ds.db.GetKeysWith(as.ds.path, func(value []byte) bool {
		if value == nil {
			return false // value does not exists or is bucket
		}
		account := &pb.Account{}
		err := proto.Unmarshal(value, account)
		if err != nil {
			return false
		}
		if account.Selected == selectType {
			return true
		}
		return false
	})
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return nil, ErrNoAccountsFound
	}
	return keys, nil
}

func (as *EthAccountStore) GetAccountsWithPrefix(n int, prefix []byte, selectType pb.Type) (Accounts, error) {
	as.lock.Lock()
	defer as.lock.Unlock()

	keys, err := as.ds.db.IndexRead(as.ds.path, prefix, n)
	if err != nil {
		return nil, err
	}

	accounts, err := as.GetAccounts(keys...)
	if err != nil {
		return nil, err
	}

	if len(accounts) == 0 || len(accounts) < n {
		return nil, ErrNoAccountsFound
	}
	return accounts, nil
}

func (as *EthAccountStore) GetAndSelect(n int, selectType pb.Type) (Accounts, error) {
	as.lock.Lock()
	defer as.lock.Unlock()

	var accounts Accounts
	if err := as.ds.db.Map(as.ds.path, n, func(value []byte) (bool, []byte, error) {
		account := &pb.Account{}
		err := proto.Unmarshal(value, account)
		if err != nil {
			return false, nil, err
		}

		if account.GetSelected() == pb.Type_NONE {
			// Select and update account
			account.Selected = selectType
			value, err = proto.Marshal(account)
			if err != nil {
				return false, nil, err
			}
			accounts = append(accounts, account)
			return true, value, nil
		}
		return false, nil, nil
	}); err != nil {
		return accounts, err
	}
	if len(accounts) == 0 || len(accounts) < n {
		return accounts, ErrNoAccountsFound
	}
	return accounts, nil
}

func (as *EthAccountStore) GetByType(n int, selectType pb.Type) (Accounts, error) {
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
			break
		}
	}
	if len(accounts) == 0 || len(accounts) < n {
		return accounts, ErrNoAccountsFound
	}
	return accounts, err
}

// All returns all accounts under the last bucket in the datastore path
func (as *EthAccountStore) All() (Accounts, error) {
	as.lock.Lock()
	defer as.lock.Unlock()

	var accounts Accounts
	err := as.ds.db.IterValues(as.ds.path, func(value []byte) error {
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

func (as *EthAccountStore) DeleteAccount(keys ...[]byte) error {
	for _, key := range keys {
		err := as.ds.db.Delete(as.ds.path, key)
		if err != nil {
			return err
		}
	}
	return nil
}
