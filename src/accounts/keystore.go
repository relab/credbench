package accounts

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/r0qs/dvcp/src/config"
	"io/ioutil"
	"os"
)

// TODO: manage multiple keys
func OpenKeystore() (*ecdsa.PrivateKey, error) {
	var account accounts.Account
	var err error

	path, err := getKeystorePath()
	password, err := getPassword()
	if err != nil {
		return nil, err
	}

	ks := keystore.NewKeyStore(path, keystore.StandardScryptN, keystore.StandardScryptP)
	if len(ks.Accounts()) == 0 {
		account, err = ks.NewAccount(password)
	} else {
		account = ks.Accounts()[0] //Return the first account
	}

	key, err := decryptKeyFile(account.URL.Path, password)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func getPassword() (string, error) {
	password := config.GetWalletConfig().Password
	//TODO look for golang library to read commands
	return password, nil
}

func getKeystorePath() (string, error) {
	path := config.GetWalletConfig().Keystore

	// Use default keystore dir if not specified
	if path == "" {
		path = config.DefaultKeyStore()
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0750); err != nil {
			return "", fmt.Errorf("cannot create keystore directory: %v", err)
		}
	}

	return path, nil
}

func decryptKeyFile(path string, password string) (*ecdsa.PrivateKey, error) {
	keyjson, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open account file: %v", err)
	}

	// FIXME: zeroes the private key in memory
	key, err := keystore.DecryptKey(keyjson, password)
	if err != nil {
		return nil, err
	}

	return key.PrivateKey, nil
}
