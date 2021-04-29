package metrics

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

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

type TXMetric struct {
	Contract string
	CAddress string
	Sender   string
	Subject  string
	Method   string
	Gas      GasMetric
}

func (u TXMetric) String() string {
	var l []string
	l = append(l, fmt.Sprintf("Sender: %s", u.Sender))
	l = append(l, fmt.Sprintf("Method: %s", u.Method))
	l = append(l, u.Gas.String())
	return strings.Join(l, "\n")
}

// Based on: https://github.com/relab/gorums/blob/master/benchmark/stats.go
type Stats struct {
	mx        sync.Mutex
	startTime time.Time
	endTime   time.Time
	startMs   runtime.MemStats
	endMs     runtime.MemStats

	count    uint64
	mean, m2 float64

	totalGas  *big.Int
	gasLimit  *big.Int
	gasPrice  *big.Int
	txMetrics chan TXMetric
}

func NewStatsTracker(gasLimit, gasPrice *big.Int) *Stats {
	return &Stats{
		gasLimit:  gasLimit,
		gasPrice:  gasPrice,
		txMetrics: make(chan TXMetric),
	}
}

func (s *Stats) StartLogger(filename string) error {
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
	for m := range s.txMetrics {
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
		"gasPrice":     s.gasPrice,
		"gasLimit":     s.gasLimit,
	}).Info()
	return nil
}

// Start records the start time and memory stats
func (s *Stats) Start() {
	s.mx.Lock()
	defer s.mx.Unlock()

	runtime.ReadMemStats(&s.startMs)
	s.startTime = time.Now()
}

// End records the end time and memory stats
func (s *Stats) End() {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.endTime = time.Now()
	runtime.ReadMemStats(&s.endMs)
	close(s.txMetrics)
}

func (s *Stats) AddTXMetric(m TXMetric) {
	s.mx.Lock()
	s.totalGas.Add(s.totalGas, m.Gas.GasUsed)
	s.mx.Unlock()

	s.txMetrics <- m
}

// AddLatency adds a latency measurement
func (s *Stats) AddLatency(l time.Duration) {
	s.mx.Lock()
	defer s.mx.Unlock()

	// implements Welford's algorithm
	s.count++
	delta := float64(l) - s.mean
	s.mean += delta / float64(s.count)
	delta2 := float64(l) - s.mean
	s.m2 += delta * delta2
}

type BenchmarkResult struct {
	TotalOps    uint64
	TotalGas    uint64
	TotalTime   int64
	Throughput  float64
	LatencyAvg  float64
	LatencyVar  float64
	AllocsPerOp uint64
	MemPerOp    uint64
}

// Format returns a tab formatted string representation of the performance metrics
func (br *BenchmarkResult) Format() string {
	latency := br.LatencyAvg / float64(time.Millisecond)
	latencySD := math.Sqrt(br.LatencyVar) / float64(time.Millisecond)
	latencyVariance := math.Pow(latencySD, 2)
	b := new(strings.Builder)
	fmt.Fprintf(b, "%.2f ops/sec\t", br.Throughput)
	fmt.Fprintf(b, "%.2f ms\t", latency)
	fmt.Fprintf(b, "%.2f ms\t", latencySD)
	fmt.Fprintf(b, "%.2f ms\t", latencyVariance)
	fmt.Fprintf(b, "%d B/op\t", br.MemPerOp)
	fmt.Fprintf(b, "%d allocs/op\t", br.AllocsPerOp)
	return b.String()
}

// GetBenchmarkResult computes and returns the results of the benchmark
func (s *Stats) GetBenchmarkResult() *BenchmarkResult {
	s.mx.Lock()
	defer s.mx.Unlock()

	br := &BenchmarkResult{}
	br.TotalOps = s.count
	br.TotalGas = s.totalGas.Uint64()
	br.TotalTime = int64(s.endTime.Sub(s.startTime))
	br.Throughput = float64(br.TotalOps) / float64(time.Duration(br.TotalTime).Seconds())
	br.LatencyAvg = s.mean
	if s.count > 2 {
		br.LatencyVar = s.m2 / float64(s.count-1)
	}
	br.AllocsPerOp = (s.endMs.Mallocs - s.startMs.Mallocs) / br.TotalOps
	br.MemPerOp = (s.endMs.TotalAlloc - s.startMs.TotalAlloc) / br.TotalOps
	return br
}

// Clear zeroes out the stats
func (s *Stats) Clear() {
	s.mx.Lock()
	s.startTime = time.Time{}
	s.endTime = time.Time{}
	s.startMs = runtime.MemStats{}
	s.endMs = runtime.MemStats{}
	s.count = 0
	s.totalGas = big.NewInt(0)
	s.mean = 0
	s.m2 = 0
	s.txMetrics = make(chan TXMetric)
	s.mx.Unlock()
}
