package transactor

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/relab/ct-eth-dapp/bench/eth"
	"github.com/relab/ct-eth-dapp/bench/metrics"
	log "github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Transactor keep gas metrics per account
type Transactor struct {
	backend *ethclient.Client
	Metrics chan metrics.UsageMetric
}

func NewTransactor(backend *ethclient.Client) *Transactor {
	return &Transactor{
		backend: backend,
		Metrics: make(chan metrics.UsageMetric),
	}
}

func (t *Transactor) Close() {
	close(t.Metrics)
}

// SendTX performs a raw transaction and collect gas metrics
func (t *Transactor) SendTX(contractName string, opts *bind.TransactOpts, contractAddress common.Address, contractABI string, method string, params ...interface{}) (*types.Transaction, error) {
	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return nil, err
	}
	input, err := parsedABI.Pack(method, params...)
	if err != nil {
		return nil, err
	}
	msg := ethereum.CallMsg{
		From:     opts.From,
		To:       &contractAddress,
		GasPrice: opts.GasPrice,
		Value:    opts.Value,
		Data:     input,
	}
	gas, err := t.backend.EstimateGas(context.TODO(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas needed: %v", err)
	}

	contract := bind.NewBoundContract(contractAddress, parsedABI, t.backend, t.backend, t.backend)

	tx, err := contract.Transact(opts, method, params...)
	if err != nil {
		return nil, err
	}

	// https://ethereum.github.io/yellowpaper/paper.pdf
	// log.Debugf("Gas Usage per execution: %d Gas\n", gas-21000) // subtract minimum transaction cost
	gasCost := eth.CalculateGasCost(gas, tx.GasPrice())
	log.Debugf("Gas Cost (ether): %v\n", eth.WeiToEther(gasCost))
	// TODO: Estimate Fiat value (USD and NOK)

	metric := metrics.UsageMetric{
		Contract: contractName,
		CAddress: contractAddress.Hex(),
		Sender:   opts.From.Hex(),
		Method:   method,
		Gas: metrics.GasMetric{
			GasUsed:    new(big.Int).SetUint64(gas),
			GasPrice:   tx.GasPrice(),
			GasCostWei: gasCost,
		},
	}

	switch method {
	case "registerCredential", "aggregateCredentials", "addStudent":
		s := params[0].(common.Address)
		metric.Subject = s.Hex()
	}

	t.Metrics <- metric

	log.Infof("Tx sent: %x\n", tx.Hash())
	return tx, nil
}

// Deploy performs a raw transaction to deploy a contract and collect gas metrics
func (t *Transactor) Deploy(opts *bind.TransactOpts, backend bind.ContractBackend, contractABI string, contractCode string, params ...interface{}) (common.Address, *types.Transaction, *bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	input, err := parsed.Pack("", params...)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	msg := ethereum.CallMsg{
		From:     opts.From,
		GasPrice: opts.GasPrice,
		Value:    opts.Value,
		Data:     input,
	}
	gas, err := t.backend.EstimateGas(context.TODO(), msg)
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, c, err := bind.DeployContract(opts, parsed, common.FromHex(contractCode), backend, params...)
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	gasCost := eth.CalculateGasCost(gas, tx.GasPrice())
	log.Debugf("Gas Cost (ether): %v\n", eth.WeiToEther(gasCost))

	metric := metrics.UsageMetric{
		Contract: parsed.Constructor.Name,
		CAddress: address.Hex(),
		Sender:   opts.From.Hex(),
		Method:   "deploy",
		Gas: metrics.GasMetric{
			GasUsed:    new(big.Int).SetUint64(gas),
			GasPrice:   tx.GasPrice(),
			GasCostWei: gasCost,
		},
	}
	t.Metrics <- metric
	return address, tx, c, nil
}
