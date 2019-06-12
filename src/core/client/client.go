package client

import (
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
)

type BBChainEthClient interface {
	Backend() (*ethclient.Client, error)
	Close()
}

type client struct {
	backend *ethclient.Client
}

func NewClient(url string) (BBChainEthClient, error) {
	c := &client{}

	backend, err := ethclient.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the Ethereum client: %v", err)
	}
	c.backend = backend
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
