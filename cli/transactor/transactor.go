package transactor

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	golog "log"

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

type TXProfiler struct {
	gasLimit *big.Int
	gasPrice *big.Int
}

func NewTXProfiler(gasLimit, gasPrice *big.Int) *TXProfiler {
	return &TXProfiler{
		gasLimit: gasLimit,
		gasPrice: gasPrice,
	}
}

func (p *TXProfiler) SaveMetric(metrics chan UsageMetric) error {
	f, err := os.OpenFile("gasMetrics_"+fmt.Sprintf("%d", time.Now().UnixNano())+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	logger := golog.New(f, "", 0)
	n := 0
	for m := range metrics {
		logger.Printf("contract:%s;address:%s;sender:%s;method:%s;gasUsed:%s;gasCost:%s", m.Contract, m.CAddress, m.Sender, m.Method, m.Gas.GasUsed.String(), m.Gas.GasCostWei.String())
		n++
	}
	logger.Printf("totalEntries:%d;gasPrice:%s;gasLimit:%s", n, p.gasPrice, p.gasLimit)
	return nil
}

type UsageMetric struct {
	Contract string
	CAddress string
	Sender   string
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

	t.Metrics <- UsageMetric{
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

	log.Infof("Tx sent: %x\n", tx.Hash())
	return tx, nil
}
