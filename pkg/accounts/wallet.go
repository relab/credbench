package accounts

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"math/big"

	ethAccounts "github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/console/prompt"
	"github.com/ethereum/go-ethereum/crypto"

	log "github.com/sirupsen/logrus"
)

type Wallet interface {
	Account() ethAccounts.Account
	PrivateKey() *ecdsa.PrivateKey
	Unlock(password string) error
	Lock() error
	Address() common.Address
	GetTxOpts(backend bind.ContractBackend) (*bind.TransactOpts, error)
}

type wallet struct {
	account    ethAccounts.Account
	privateKey *ecdsa.PrivateKey
	keyStore   *keystore.KeyStore
	unlocked   bool
	chainID    *big.Int
}

func NewWallet(accountAddr common.Address, keystoreDir string, chainID *big.Int) (Wallet, error) {
	var account ethAccounts.Account
	var err error

	keyStore := keystore.NewKeyStore(keystoreDir, keystore.StandardScryptN, keystore.StandardScryptP)

	if len(keyStore.Accounts()) == 0 || ((accountAddr != common.Address{}) && !keyStore.HasAddress(accountAddr)) {
		// Account does not exist
		account, err = createAccount(keyStore)
		if err != nil {
			return nil, err
		}
	} else {
		account, err = getAccount(accountAddr, keyStore)
		if err != nil {
			return nil, err
		}
	}

	password, _ := getPassword(false)
	key, err := decryptKeyFile(account.URL.Path, password)
	if err != nil {
		return nil, err
	}

	return &wallet{
		account:    account,
		privateKey: key,
		keyStore:   keyStore,
		unlocked:   false,
		chainID:    chainID,
	}, nil
}

func (w *wallet) GetTxOpts(backend bind.ContractBackend) (*bind.TransactOpts, error) {
	gasPrice, err := backend.SuggestGasPrice(context.TODO())
	if err != nil {
		return nil, err
	}

	nonce, err := backend.PendingNonceAt(context.TODO(), w.account.Address)
	if err != nil {
		return nil, err
	}

	// EIP155 replay protected TX
	// https://github.com/ethereum/go-ethereum/pull/22339
	transactOpts, err := bind.NewKeyedTransactorWithChainID(w.privateKey, w.chainID)
	if err != nil {
		return nil, err
	}
	transactOpts.GasLimit = uint64(6721975) // FIXME: get from config file
	transactOpts.GasPrice = gasPrice

	transactOpts.Nonce = new(big.Int).SetUint64(nonce)
	return transactOpts, nil
}

func (w *wallet) Unlock(password string) (err error) {
	err = w.keyStore.Unlock(w.account, password)
	if err != nil {
		if password != "" {
			return err
		}
		log.Infof("Please enter the password to unlock Ethereum account %v:", w.account.Address.Hex())

		password, err = getPassword(false)
		if err != nil {
			return err
		}

		err = w.keyStore.Unlock(w.account, password)
		if err != nil {
			return err
		}
	}
	w.unlocked = true

	log.Infof("Unlocked ETH account: %v\n", w.account.Address.Hex())
	return nil
}

func (w *wallet) Lock() error {
	err := w.keyStore.Lock(w.account.Address)
	if err != nil {
		return err
	}
	w.unlocked = false

	return nil
}

func (w *wallet) Account() ethAccounts.Account {
	return w.account
}

func (w *wallet) Address() common.Address {
	return w.account.Address
}

func (w *wallet) HexAddress() string {
	return w.Address().Hex()
}

func (w *wallet) PrivateKey() *ecdsa.PrivateKey {
	return w.privateKey
}

func ImportKey(hexkey string, keyStore *keystore.KeyStore) (Wallet, error) {
	key, err := crypto.HexToECDSA(hexkey)
	if err != nil {
		return nil, fmt.Errorf("Error parsing the private key: %v", err)
	}

	password, _ := getPassword(true)
	account, err := keyStore.ImportECDSA(key, password)
	if err != nil {
		return nil, fmt.Errorf("Error importing the private key: %v", err)
	}

	log.Infof("Account address %v successfully imported\n", account.Address.Hex())
	return &wallet{
		account:    account,
		privateKey: key,
		keyStore:   keyStore,
		unlocked:   true,
	}, nil
}

func NewAccount(keystoreDir string) (err error) {
	var account ethAccounts.Account

	keyStore := keystore.NewKeyStore(keystoreDir, keystore.StandardScryptN, keystore.StandardScryptP)
	account, err = createAccount(keyStore)
	if err != nil {
		return err
	}
	log.Infof("Account created: %v\n", account.Address.Hex())
	return nil
}

func GetAccountAddress(addr string, keystoreDir string) common.Address {
	var account ethAccounts.Account

	address := common.HexToAddress(addr)
	keyStore := keystore.NewKeyStore(keystoreDir, keystore.StandardScryptN, keystore.StandardScryptP)
	if (address != common.Address{}) && keyStore.HasAddress(address) {
		return address
	}
	account, _ = getAccount(address, keyStore)
	address = account.Address
	return address
}

func GetAddress(key *ecdsa.PrivateKey) common.Address {
	return crypto.PubkeyToAddress(key.PublicKey)
}

func getAccount(accountAddr common.Address, keyStore *keystore.KeyStore) (ethAccounts.Account, error) {
	accounts := keyStore.Accounts()
	if len(accounts) == 0 {
		return ethAccounts.Account{}, fmt.Errorf("no accounts, please create one")
	}

	if (accountAddr != common.Address{}) {
		for _, account := range accounts {
			if account.Address == accountAddr {
				return account, nil
			}
		}

		return ethAccounts.Account{}, fmt.Errorf("Ethereum account not found")
	}
	// Return the first existent account
	return accounts[0], nil
}

func createAccount(keyStore *keystore.KeyStore) (ethAccounts.Account, error) {
	log.Infoln("Creating a new Ethereum account")
	password, err := getPassword(true)
	if err != nil {
		return ethAccounts.Account{}, err
	}

	return keyStore.NewAccount(password)
}

func getPassword(repeat bool) (string, error) {
	password, err := prompt.Stdin.PromptPassword("Password: ")
	if err != nil {
		return "", err
	}

	if repeat {
		confirmation, err := prompt.Stdin.PromptPassword("Repeat password: ")
		if err != nil {
			return "", err
		}

		if password != confirmation {
			return "", fmt.Errorf("passwords do not match")
		}
	}

	return password, nil
}

func decryptKeyFile(path string, password string) (*ecdsa.PrivateKey, error) {
	keyJSON, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open account file: %v", err)
	}

	key, err := keystore.DecryptKey(keyJSON, password)
	if err != nil {
		return nil, err
	}

	return key.PrivateKey, nil
}
