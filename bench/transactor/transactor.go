package transactor

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type GasMetric struct {
	GasUsed    *big.Int
	GasPrice   *big.Int
	GasCostWei *big.Int
}

func (g GasMetric) String() string {
	var l []string
	l = append(l, fmt.Sprintf("Gas Usage: %v gas", g.GasUsed.String()))
	l = append(l, fmt.Sprintf("Gas Price: %v gas", g.GasPrice.String()))
	l = append(l, fmt.Sprintf("Gas Cost (ether): %v", WeiToEther(g.GasCostWei).String()))
	return strings.Join(l, "\n")
}

type LogEntryFormatter struct{}

func (*LogEntryFormatter) Format(entry *log.Entry) ([]byte, error) {
	b, err := json.Marshal(entry.Data)
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

type TXLogger struct {
	gasLimit *big.Int
	gasPrice *big.Int
}

func NewTXLogger(gasLimit, gasPrice *big.Int) *TXLogger {
	return &TXLogger{
		gasLimit: gasLimit,
		gasPrice: gasPrice,
	}
}

// add period on the metric file
func (p *TXLogger) SaveMetric(filename string, metrics chan UsageMetric) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	logger := log.New()
	logger.SetOutput(f)
	logger.SetLevel(log.InfoLevel)
	logger.SetFormatter(new(LogEntryFormatter))
	var n uint = 0
	for m := range metrics {
		entry := log.Fields{
			"contract": m.Contract,
			"address":  m.CAddress,
			"sender":   m.Sender,
			"method":   m.Method,
			"gasUsed":  m.Gas.GasUsed,
			"gasCost":  m.Gas.GasCostWei,
		}
		if len(m.Subject) > 0 {
			entry["subject"] = m.Subject
		}
		logger.WithFields(entry).Info()
		n++
	}
	logger.WithFields(log.Fields{
		"totalEntries": n,
		"gasPrice":     p.gasPrice,
		"gasLimit":     p.gasLimit,
	}).Info()
	return nil
}

type UsageMetric struct {
	Contract string
	CAddress string
	Sender   string
	Subject  string
	Method   string
	Gas      GasMetric
}

func (u UsageMetric) String() string {
	var l []string
	l = append(l, fmt.Sprintf("Sender: %s", u.Sender))
	l = append(l, fmt.Sprintf("Method: %s", u.Method))
	l = append(l, u.Gas.String())
	return strings.Join(l, "\n")
}

// Transactor keep gas metrics per account
type Transactor struct {
	backend *ethclient.Client
	Metrics chan UsageMetric
}

func NewTransactor(backend *ethclient.Client) *Transactor {
	return &Transactor{
		backend: backend,
		Metrics: make(chan UsageMetric),
	}
}

func (t *Transactor) Close() {
	close(t.Metrics)
}

// SendTX performs a transaction and collect gas metrics
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
	gasCost := CalculateGasCost(gas, tx.GasPrice())
	log.Debugf("Gas Cost (ether): %v\n", WeiToEther(gasCost))
	// TODO: Estimate Fiat value (USD and NOK)

	metric := UsageMetric{
		Contract: contractName,
		CAddress: contractAddress.Hex(),
		Sender:   opts.From.Hex(),
		Method:   method,
		Gas: GasMetric{
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
		From: opts.From,
		// To:       &contractAddress,
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

	gasCost := CalculateGasCost(gas, tx.GasPrice())
	log.Debugf("Gas Cost (ether): %v\n", WeiToEther(gasCost))

	metric := UsageMetric{
		Contract: parsed.Constructor.Name,
		CAddress: address.Hex(),
		Sender:   opts.From.Hex(),
		Method:   "deploy",
		Gas: GasMetric{
			GasUsed:    new(big.Int).SetUint64(gas),
			GasPrice:   tx.GasPrice(),
			GasCostWei: gasCost,
		},
	}
	t.Metrics <- metric
	return address, tx, c, nil
}
