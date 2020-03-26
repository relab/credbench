package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/spf13/cobra"

	"github.com/relab/bbchain-dapp/src/core/accounts"
	"github.com/relab/bbchain-dapp/src/core/client"
	"github.com/relab/bbchain-dapp/src/utils"
)

var (
	keyStore   *keystore.KeyStore
	wallet     accounts.BBChainWallet
	senderAddr common.Address
	clientConn client.BBChainEthClient
)

func setupPreRun() (err error) {
	err = utils.CreateDir(datadir)
	if err != nil {
		return err
	}
	return nil
}

func setupClient() (err error) {
	err = loadWallet()
	if err != nil {
		return err
	}

	clientConn, err = newClientConn()
	if err != nil {
		return err
	}

	if waitPeers {
		err = checkConnectPeers()
		if err != nil {
			return err
		}
	}

	return nil
}

func clientClose(_ *cobra.Command, _ []string) {
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

func loadWallet() (err error) {
	senderAddr = accounts.GetAccountAddress(defaultAccount, keystoreDir)
	wallet, err = accounts.NewWallet(senderAddr, keystoreDir)
	return err
}

func loadKeystore() {
	keyStore = keystore.NewKeyStore(keystoreDir, keystore.StandardScryptN, keystore.StandardScryptP)
}

func checkConnectPeers() error {
	client, err := rpc.DialIPC(context.Background(), ipcFile)
	if err != nil {
		return err
	}
	var peers []*p2p.PeerInfo
	client.Call(&peers, "admin_peers")

	start := time.Now()
	for len(peers) < 1 { //TODO: get number of peers as parameter
		fmt.Printf("%v peers connected. Waiting for peers...\n", len(peers))
		t := time.Now()
		elapsed := t.Sub(start)
		if elapsed > defaultWaitTime {
			return fmt.Errorf("timeout waiting for peers after %v seconds", elapsed)
		}
		time.Sleep(1 * time.Second)
		client.Call(&peers, "admin_peers")
	}
	fmt.Printf("Connected to %v peers.", len(peers))
	return nil
}
