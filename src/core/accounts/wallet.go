package accounts

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"

	ethAccounts "github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/console/prompt"
	"github.com/ethereum/go-ethereum/crypto"
)

// TODO: separate account and wallet
type BBChainWallet interface {
	Account() ethAccounts.Account
	PrivateKey() *ecdsa.PrivateKey
	Unlock(password string) error
	Lock() error
}

type wallet struct {
	account    ethAccounts.Account
	privateKey *ecdsa.PrivateKey
	keyStore   *keystore.KeyStore
	unlocked   bool
}

func NewWallet(accountAddr common.Address, keystoreDir string) (BBChainWallet, error) {
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

	fmt.Printf("Using Ethereum account: %v\n", account.Address.Hex())

	return &wallet{
		account:    account,
		privateKey: key,
		keyStore:   keyStore,
		unlocked:   false,
	}, nil
}

func ImportKey(hexkey string, keyStore *keystore.KeyStore) (BBChainWallet, error) {
	key, err := crypto.HexToECDSA(hexkey)
	if err != nil {
		return nil, fmt.Errorf("Error parsing the private key: %v", err)
	}

	password, _ := getPassword(true)
	account, err := keyStore.ImportECDSA(key, password)

	fmt.Printf("Account address %v successfully imported\n", account.Address.Hex())
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
	fmt.Printf("Account created: %v\n", account.Address.Hex())
	return nil
}

func GetTxOpts(key *ecdsa.PrivateKey, backend bind.ContractBackend) (*bind.TransactOpts, error) {
	gasPrice, err := backend.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("Failed to estimate the gas price: %v", err)
	}
	transactOpts := bind.NewKeyedTransactor(key)
	transactOpts.GasLimit = uint64(6721975) //TODO: get from config file
	transactOpts.GasPrice = gasPrice
	return transactOpts, nil
}

func (w *wallet) Unlock(password string) (err error) {
	err = w.keyStore.Unlock(w.account, password)
	if err != nil {
		if password != "" {
			return err
		}
		fmt.Printf("Please enter the password to unlock Ethereum account %v:", w.account.Address.Hex())

		password, err = getPassword(false)
		err = w.keyStore.Unlock(w.account, password)
		if err != nil {
			return err
		}
	}
	w.unlocked = true

	fmt.Printf("Unlocked ETH account: %v\n", w.account.Address.Hex())
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

func (w *wallet) PrivateKey() *ecdsa.PrivateKey {
	return w.privateKey
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
	//Return the first existent account
	return accounts[0], nil
}

func createAccount(keyStore *keystore.KeyStore) (ethAccounts.Account, error) {
	fmt.Println("Creating a new Ethereum account")
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

func GetKeys(hexkey string) (*ecdsa.PrivateKey, common.Address, error) {
	key, err := crypto.HexToECDSA(hexkey)
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("Error parsing the private key: %v", err)
	}
	address := crypto.PubkeyToAddress(key.PublicKey)
	return key, address, nil
}
