package accounts

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"sync"
	"sync/atomic"

	ethAccounts "github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// Wallet used for test purposes only
type genWallet struct {
	lock    sync.Mutex
	key     *ecdsa.PrivateKey
	address common.Address
	nonce   uint64 // nonce is a local atomic transaction counter
}

func NewGenWallet(accountAddr common.Address, accountHexKey string) Wallet {
	return &genWallet{
		key:     HexToKey(accountHexKey),
		address: accountAddr,
	}
}

func (w *genWallet) GetTxOpts(backend bind.ContractBackend) (*bind.TransactOpts, error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	gasPrice, err := backend.SuggestGasPrice(context.TODO())
	if err != nil {
		return nil, err
	}

	nonce, err := backend.PendingNonceAt(context.TODO(), w.address)
	if err != nil {
		return nil, err
	}
	if w.nonce < nonce {
		w.nonce = nonce
	}

	transactOpts := bind.NewKeyedTransactor(w.key)
	transactOpts.GasLimit = uint64(6721975) // FIXME: get from config file
	transactOpts.GasPrice = gasPrice
	// Note: overflow may happen when converting uint64 to int64
	transactOpts.Nonce = new(big.Int).SetUint64(w.nonce)

	w.IncNonce()
	return transactOpts, nil
}

func (w *genWallet) IncNonce() uint64 {
	return atomic.AddUint64(&w.nonce, 1)
}

func (w *genWallet) GetNextNonce() uint64 {
	return w.IncNonce()
}

func (w *genWallet) Unlock(password string) (err error) {
	return nil // Not implemented
}

func (w *genWallet) Lock() error {
	return nil // Not implemented
}

func (w *genWallet) Account() ethAccounts.Account {
	return ethAccounts.Account{Address: w.address}
}

func (w *genWallet) Address() common.Address {
	return w.address
}

func (w *genWallet) PrivateKey() *ecdsa.PrivateKey {
	return w.key
}
