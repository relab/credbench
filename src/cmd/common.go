package cmd

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/relab/bbchain-dapp/src/core/accounts"
	"github.com/relab/bbchain-dapp/src/core/client"
	"github.com/spf13/cobra"
)

var (
	keyStore   *keystore.KeyStore
	wallet     accounts.BBChainWallet
	senderAddr common.Address
	clientConn client.BBChainEthClient
)

func clientPreRunE(cmd *cobra.Command, args []string) error {
	var err error

	err = loadWallet(cmd, args)
	if err != nil {
		return err
	}

	clientConn, err = newClientConn()
	if err != nil {
		return err
	}

	return nil
}

func clientPostRun(_ *cobra.Command, _ []string) {
	clientConn.Close()
}

func newClientConn() (client.BBChainEthClient, error) {
	cli, err := client.NewClient(backendURL)
	if err != nil {
		return nil, err
	}
	clientConn = cli
	return cli, err
}

func loadWallet(cmd *cobra.Command, args []string) error {
	var err error

	senderAddr = accounts.GetAccountAddress(accountAddress, keystoreDir)
	wallet, err = accounts.NewWallet(senderAddr, keystoreDir)
	return err
}

func loadKeystore(cmd *cobra.Command, args []string) {
	keyStore = keystore.NewKeyStore(keystoreDir, keystore.StandardScryptN, keystore.StandardScryptP)
}

// FIXME: remove this
func getKeys(hexkey string) (*ecdsa.PrivateKey, common.Address) {
	key, err := crypto.HexToECDSA(hexkey)
	if err != nil {
		fmt.Printf("Error parsing the private key: %v", err)
	}
	address := crypto.PubkeyToAddress(key.PublicKey)
	return key, address
}
