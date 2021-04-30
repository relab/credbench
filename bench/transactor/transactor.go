package transactor

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/relab/ct-eth-dapp/bench/eth"
	"github.com/relab/ct-eth-dapp/bench/metrics"
	"github.com/relab/ct-eth-dapp/pkg/deployer"
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
	Stats   *metrics.Stats
}

func NewTransactor(backend *ethclient.Client, gasLimit, gasPrice *big.Int) *Transactor {
	return &Transactor{
		backend: backend,
		Stats:   metrics.NewStatsTracker(gasLimit, gasPrice),
	}
}

// SendTX performs a raw transaction and collect gas metrics
func (t *Transactor) SendTX(contractName string, opts *bind.TransactOpts, contractAddress common.Address, contractABI string, method string, params ...interface{}) (*types.Transaction, error) {
	sendTime := time.Now().UnixNano()

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
	gasLimit, err := t.backend.EstimateGas(context.TODO(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas needed: %v", err)
	}

	contract := bind.NewBoundContract(contractAddress, parsedABI, t.backend, t.backend, t.backend)

	tx, err := contract.Transact(opts, method, params...)
	if err != nil {
		return nil, err
	}

	// async wait for tx confirmation
	go func() {
		receipt, err := deployer.WaitTxReceipt(context.TODO(), t.backend, tx, 1*time.Minute)
		if err != nil {
			log.Errorf("Execution error in tx %x: %v\n", tx.Hash(), err)
			return
		}
		if gasLimit != receipt.GasUsed {
			log.Warnf("Gas estimation differs | estimated: %v used:%v\n", gasLimit, receipt.GasUsed)
		}
		if receipt.Status == types.ReceiptStatusFailed {
			log.Errorf("Tx %x execution failed with receipt logs: %v\n", tx.Hash(), receipt.Logs)
		}
		// we compute the performance metrics of any executed transactions
		latency := time.Now().UnixNano() - sendTime
		t.Stats.AddExecMetric(time.Duration(latency), new(big.Int).SetUint64(gasLimit))
		// https://ethereum.github.io/yellowpaper/paper.pdf
		// log.Debugf("Gas Usage per execution: %d Gas\n", gas-21000) // subtract minimum transaction cost
		gasCost := eth.CalculateGasCost(gasLimit, tx.GasPrice())
		log.Debugf("Gas Cost (ether): %v\n", eth.WeiToEther(gasCost))
		// TODO: Estimate Fiat value (USD and NOK)

		metric := metrics.TXMetric{
			Contract: contractName,
			CAddress: contractAddress.Hex(),
			Sender:   opts.From.Hex(),
			Method:   method,
			Gas: metrics.GasMetric{
				GasUsed:    new(big.Int).SetUint64(gasLimit),
				GasPrice:   tx.GasPrice(),
				GasCostWei: gasCost,
			},
			Latency: latency,
		}

		switch method {
		case "registerCredential", "aggregateCredentials", "addStudent":
			s := params[0].(common.Address)
			metric.Subject = s.Hex()
		}
		t.Stats.AddTXMetric(metric)
	}()

	log.Infof("Tx sent: %x\n", tx.Hash())
	return tx, nil
}

// Deploy performs a raw transaction to deploy a contract and collect gas metrics
func (t *Transactor) Deploy(opts *bind.TransactOpts, backend bind.ContractBackend, contractABI string, contractCode string, params ...interface{}) (common.Address, *types.Transaction, *bind.BoundContract, error) {
	sendTime := time.Now().UnixNano()

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
	gasLimit, err := t.backend.EstimateGas(context.TODO(), msg)
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, c, err := bind.DeployContract(opts, parsed, common.FromHex(contractCode), backend, params...)
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	go func() {
		receipt, err := deployer.WaitTxReceipt(context.TODO(), t.backend, tx, 1*time.Minute)
		if err != nil {
			log.Errorf("Execution error in tx %x: %v\n", tx.Hash(), err)
			return
		}
		if gasLimit != receipt.GasUsed {
			log.Warnf("Gas estimation differs | estimated: %v used:%v\n", gasLimit, receipt.GasUsed)
		}
		if receipt.Status == types.ReceiptStatusFailed {
			log.Errorf("Tx %x execution failed with receipt logs: %v\n", tx.Hash(), receipt.Logs)
		}

		latency := time.Now().UnixNano() - sendTime
		t.Stats.AddExecMetric(time.Duration(latency), new(big.Int).SetUint64(gasLimit))

		gasCost := eth.CalculateGasCost(gasLimit, tx.GasPrice())
		log.Debugf("Gas Cost (ether): %v\n", eth.WeiToEther(gasCost))

		metric := metrics.TXMetric{
			Contract: parsed.Constructor.Name,
			CAddress: address.Hex(),
			Sender:   opts.From.Hex(),
			Method:   "deploy",
			Gas: metrics.GasMetric{
				GasUsed:    new(big.Int).SetUint64(gasLimit),
				GasPrice:   tx.GasPrice(),
				GasCostWei: gasCost,
			},
			Latency: latency,
		}

		t.Stats.AddTXMetric(metric)
	}()

	return address, tx, c, nil
}
