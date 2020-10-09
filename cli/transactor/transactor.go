package transactor

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	keyutils "github.com/relab/ct-eth-dapp/src/accounts"
)

type GasMetric struct {
	gasUsed    *big.Int
	gasPrice   *big.Int
	gasCostWei *big.Int
}

type Transactor struct {
	backend       *ethclient.Client
	gasLimit      *big.Int
	methods       map[string]GasMetric
	totalGasUsage GasMetric
	quit          chan struct{}
}

func NewTransactor(backend *ethclient.Client) *Transactor {
	return &Transactor{
		backend: backend,
		quit:    make(chan struct{}),
	}
}

// TODO implement caller
func (t Transactor) SendTX(senderHexKey string, contractAddress common.Address, contractABI string, method string, params ...interface{}) (*types.Transaction, error) {
	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return nil, err
	}

	input, err := parsedABI.Pack(method, params...)
	if err != nil {
		return nil, err
	}
	opts, _ := keyutils.GetTxOpts(keyutils.HexToKey(senderHexKey), t.backend)
	msg := ethereum.CallMsg{
		From:     opts.From,
		To:       &contractAddress,
		GasPrice: opts.GasPrice,
		Value:    opts.Value,
		Data:     input,
	}
	gas, err := t.backend.EstimateGas(context.Background(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas needed: %v", err)
	}

	// TODO: store gas usage
	fmt.Printf("Gas Usage: %v gas\n", gas)
	// https://ethereum.github.io/yellowpaper/paper.pdf
	fmt.Printf("Gas Usage per execution: %d Gas\n", gas-21000) // subtract minimum transaction cost
	gasPrice, _ := new(big.Int).SetString("20000000000", 10)
	fmt.Printf("Gas Price: %v\n", gasPrice)
	gasCost := CalculateGasCost(gas, gasPrice)
	fmt.Printf("Gas Cost (wei): %v\n", gasCost)
	fmt.Printf("Gas Cost (ether): %v\n", WeiToEther(gasCost))
	//TODO: Estimate Fiat value (USD and NOK)

	fmt.Printf("Calling method: %s\n", method)
	contract := bind.NewBoundContract(contractAddress, parsedABI, t.backend, t.backend, t.backend)
	tx, err := contract.Transact(opts, method, params...)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Tx sent: %x\n", tx.Hash())
	return tx, nil
}
