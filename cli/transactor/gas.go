package transactor

import (
	"context"
	"log"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// CalculateGasCost given gas limit (units) and gas price (wei)
func CalculateGasCost(gasLimit uint64, gasPrice *big.Int) *big.Int {
	gasLimitBig := big.NewInt(int64(gasLimit))
	return gasLimitBig.Mul(gasLimitBig, gasPrice)
}

// WeiToEther converts wei unit to ether
func WeiToEther(wei *big.Int) *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(math.Pow10(18)))
}

func GetBalance(hexAddress string, backend *ethclient.Client) *big.Float {
	address := common.HexToAddress(hexAddress)
	balance, err := backend.BalanceAt(context.Background(), address, nil)
	if err != nil {
		log.Fatal(err)
	}
	return WeiToEther(balance)
}
