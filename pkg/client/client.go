package client

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
)

type EthClient interface {
	Backend() (*ethclient.Client, error)
	Close()
	CheckConnectPeers(timeout time.Duration) error
}

// TODO: add fixed embedded contract sessions and addresses
type client struct {
	url     string
	rpc     *rpc.Client
	backend *ethclient.Client
}

func NewClient(url string) (EthClient, error) {
	c := &client{}

	rpcc, err := rpc.DialContext(context.Background(), url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the Ethereum client: %v", err)
	}
	c.backend = ethclient.NewClient(rpcc)
	c.rpc = rpcc
	c.url = url
	return c, nil
}

func (c *client) Close() {
	c.backend.Close()
}

func (c *client) Backend() (*ethclient.Client, error) {
	if c.backend == nil {
		return nil, fmt.Errorf("missing Ethereum client backend")
	}
	return c.backend, nil
}

func (c *client) CheckConnectPeers(timeout time.Duration) error {
	var peers []*p2p.PeerInfo
	err := c.rpc.Call(&peers, "admin_peers")
	if err != nil {
		return err
	}

	start := time.Now()
	for len(peers) < 1 { // TODO: get minimum number of peers as parameter
		log.Printf("%v peers connected. Waiting for peers...\n", len(peers))
		t := time.Now()
		elapsed := t.Sub(start)
		if elapsed > timeout {
			return fmt.Errorf("timeout waiting for peers after %v seconds", elapsed)
		}
		time.Sleep(1 * time.Second)
		err = c.rpc.Call(&peers, "admin_peers")
		if err != nil {
			return err
		}
	}
	log.Printf("Connected to %v peers.", len(peers))
	return nil
}
