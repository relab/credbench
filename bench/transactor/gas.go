package transactor

import (
	"context"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	log "github.com/sirupsen/logrus"
)

// CalculateGasCost given gas limit (units) and gas price (wei)
func CalculateGasCost(gasLimit uint64, gasPrice *big.Int) *big.Int {
	return new(big.Int).Mul(gasPrice, new(big.Int).SetUint64(gasLimit))
}

// WeiToEther converts wei unit to ether
func WeiToEther(wei *big.Int) *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(math.Pow10(18)))
}

func GetBalance(address common.Address, backend *ethclient.Client) *big.Float {
	balance, err := backend.BalanceAt(context.TODO(), address, nil)
	if err != nil {
		log.Fatal(err)
	}
	return WeiToEther(balance)
}
