// License Notice
// The waiting methods:
// - WaitTxReceipt
// - WaitTxConfirmation
// - WaitTxConfirmationAndFee
// were copied and adapted from:
// https://github.com/loomnetwork/go-loom/blob/97b00e86fce0e0447bcf06abdcb3ac3338c69366/client/helpers.go#L57
// And are under BSD-3-Clause License, copyright 2018 Loom Network Inc
// https://github.com/loomnetwork/go-loom/blob/97b00e86fce0e0447bcf06abdcb3ac3338c69366/License.md

package deployer

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
)

var ErrTxFailed = errors.New("transaction failed")

// WaitTxReceipt waits for a tx to be mined.
// It stops waiting if the context is canceled or the tx hasn't been confirmed after the specified timeout.
func WaitTxReceipt(ctx context.Context, client *ethclient.Client, tx *types.Transaction, timeout time.Duration) (*types.Receipt, error) {
	if timeout == 0 {
		timeout = 1 * time.Minute
	}

	timer := time.NewTimer(timeout)
	ticker := time.NewTicker(time.Second)

	defer timer.Stop()
	defer ticker.Stop()

	for {
		receipt, err := client.TransactionReceipt(ctx, tx.Hash())
		if err != nil && err != ethereum.NotFound {
			return nil, errors.Wrap(err, "failed to retrieve tx receipt")
		}
		if receipt != nil {
			return receipt, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timer.C:
			return nil, errors.New("timed out waiting for tx receipt")
		case <-ticker.C:
		}
	}
}

// WaitTxConfirmation waits for a tx to be confirmed.
func WaitTxConfirmation(ctx context.Context, client *ethclient.Client, tx *types.Transaction, timeout time.Duration) error {
	r, err := WaitTxReceipt(ctx, client, tx, timeout)
	if err != nil {
		gasPrice := new(big.Int)
		if tx.GasPrice() != nil {
			gasPrice = tx.GasPrice()
		}
		cost := new(big.Int)
		if tx.Cost() != nil {
			cost = tx.Cost()
		}
		return errors.Wrap(err,
			fmt.Sprintf(
				"tx failed (gas: %v, gasPrice: %s, cost: %s)",
				tx.Gas(), gasPrice.String(), cost.String(),
			),
		)
	}
	if r.Status != types.ReceiptStatusSuccessful {
		return ErrTxFailed
	}
	return nil
}

// WaitTxConfirmationAndFee waits for a tx to be confirmed as successful, and returns the fee paid for the tx.
func WaitTxConfirmationAndFee(ctx context.Context, client *ethclient.Client, tx *types.Transaction, timeout time.Duration) (*big.Int, error) {
	r, err := WaitTxReceipt(ctx, client, tx, timeout)
	if err != nil {
		return nil, err
	}
	if r.Status != types.ReceiptStatusSuccessful {
		return nil, ErrTxFailed
	}
	return new(big.Int).Mul(tx.GasPrice(), big.NewInt(0).SetUint64(r.GasUsed)), nil
}

// GetCallResponse returns the response message after calling a method
func GetCallResponse(client *ethclient.Client, hash common.Hash) (string, error) {
	tx, _, err := client.TransactionByHash(context.TODO(), hash)
	if err != nil {
		return "", err
	}

	from, err := types.Sender(types.NewEIP155Signer(tx.ChainId()), tx)
	if err != nil {
		return "", err
	}

	msg := ethereum.CallMsg{
		From:     from,
		To:       tx.To(),
		Gas:      tx.Gas(),
		GasPrice: tx.GasPrice(),
		Value:    tx.Value(),
		Data:     tx.Data(),
	}

	res, err := client.CallContract(context.TODO(), msg, nil)
	if err != nil {
		return "", err
	}

	return string(res), nil
}
