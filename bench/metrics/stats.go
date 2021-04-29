package metrics

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/relab/ct-eth-dapp/bench/eth"
	log "github.com/sirupsen/logrus"
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
	l = append(l, fmt.Sprintf("Gas Cost (ether): %v", eth.WeiToEther(g.GasCostWei).String()))
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
