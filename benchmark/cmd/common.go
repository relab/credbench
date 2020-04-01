package cmd

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"log"
	"path"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"

	bolt "go.etcd.io/bbolt"

	"github.com/relab/bbchain-dapp/src/core/client"

	"github.com/relab/bbchain-dapp/benchmark/database"
)

func setupClient() (err error) {
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

func setupDB(dbpath, dbfile string) (err error) {
	dbFileName := path.Join(dbpath, dbfile)
	db, err = database.NewDatabase(dbFileName, &bolt.Options{Timeout: 1 * time.Second})
	return err
}

func newClientConn() (client.BBChainEthClient, error) {
	cli, err := client.NewClient(backendURL)
	if err != nil {
		return nil, err
	}
	clientConn = cli
	return cli, err
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

// TODO: cleanup
func Has0xPrefix(input string) bool {
	return len(input) >= 2 && input[0] == '0' && (input[1] == 'x' || input[1] == 'X')
}

func keyToHex(privateKey *ecdsa.PrivateKey) string {
	keyBytes := crypto.FromECDSA(privateKey)
	return hexutil.Encode(keyBytes)
}

func hexToKey(hexkey string) *ecdsa.PrivateKey {
	if Has0xPrefix(hexkey) {
		hexkey = hexkey[2:]
	}
	key, err := crypto.HexToECDSA(hexkey)
	if err != nil {
		log.Panic(err)
	}
	return key
}

func newKey() (*ecdsa.PrivateKey, common.Address) {
	privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	return privateKey, address
}

func GetTxOpts(hexkey string) (*bind.TransactOpts, error) {
	backend, _ := clientConn.Backend()
	gasPrice, err := backend.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("Failed to estimate the gas price: %v", err)
	}
	opts := bind.NewKeyedTransactor(hexToKey(hexkey))
	opts.GasLimit = uint64(6721975)
	opts.GasPrice = gasPrice
	return opts, nil
}

func makePathList(s string) (l []string) {
	if len(s) < 0 {
		return l
	}
	s = strings.ReplaceAll(strings.TrimSpace(s), " ", "")
	for _, e := range strings.Split(s, "/") {
		if e != "" {
			l = append(l, e)
		}
	}
	return l
}
