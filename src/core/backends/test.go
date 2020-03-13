package backends

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
)

// Account represents a test account
type Account struct {
	Key     *ecdsa.PrivateKey
	Address common.Address
}

type Accounts []Account

func (accs Accounts) Addresses() []common.Address {
	accounts := make([]common.Address, 0)
	for _, acc := range accs {
		accounts = append(accounts, acc.Address)
	}
	return accounts
}

var TestAccounts Accounts

// Ganache private hex string keys
var defaultHexkeys = []string{
	"4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d",
	"6cbed15c793ce57650b9877cf6fa156fbef513c4e6134f022a85b1ffdd59b2a1",
	"6370fd033278c143179d81c5526140625662b8daa446c22ee2d73db3707e620c",
	"646f1ce2fdad0e6deeeb5c7e8e5543bdde65e86029e2fd9fc169899c440a7913",
	"add53f9a7e588d003326d1cbf9e4a43c061aadd9bc938c843a79e7b4fd2ad743",
	"395df67f0c2d2d9fe1ad08d1bc8b6627011959b79c53d7dd6a3536a33ab8a4fd",
	"e485d098507f54e7733a205420dfddbe58db035fa577fc294ebd14db90767a52",
	"a453611d9419d0e56f499079478fd72c37b251a94bfde4d19872c44cf65386e3",
	"829e924fdf021ba3dbbc4225edfece9aca04b929d6e75613329ca6f1d31c0bb4",
	"b0057716d5917badaf911b193b12b910811c1497b5bada8d7711f758981c3773",
}

// TestBackend extends the SimulatedBackend to allow
// easily testing contracts.
type TestBackend struct {
	*backends.SimulatedBackend
}

func init() {
	var accounts []Account
	for _, acc := range defaultHexkeys {
		key, addr := getKeys(acc)
		accounts = append(accounts, Account{Key: key, Address: addr})
	}
	TestAccounts = accounts
}

func NewTestBackend() *TestBackend {
	ethAccounts := make(core.GenesisAlloc)
	for _, acc := range TestAccounts {
		ethAccounts[acc.Address] = core.GenesisAccount{Balance: big.NewInt(1000000000)}
	}
	backend := backends.NewSimulatedBackend(ethAccounts, 10000000)
	return &TestBackend{backend}
}

func getKeys(hexkey string) (*ecdsa.PrivateKey, common.Address) {
	key, err := crypto.HexToECDSA(hexkey)
	if err != nil {
		log.Fatal(err)
	}
	address := crypto.PubkeyToAddress(key.PublicKey)
	return key, address
}

// duration in seconds
func (b *TestBackend) GetPeriod(duration uint64) (*big.Int, *big.Int) {
	header, _ := b.HeaderByNumber(context.Background(), nil)
	// Every backend.Commit() increases the block time in 10 secs
	// so we calculate the start time to in the next block
	startingTime := header.Time + 10
	endingTime := startingTime + duration
	return new(big.Int).SetUint64(startingTime), new(big.Int).SetUint64(endingTime)
}

func (b *TestBackend) IncreaseTime(duration time.Duration) error {
	err := b.AdjustTime(duration)
	if err != nil {
		return err
	}
	b.Commit()
	return nil
}

func (b *TestBackend) GetTxOpts(key *ecdsa.PrivateKey) (*bind.TransactOpts, error) {
	gasPrice, err := b.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("Failed to estimate the gas price: %v", err)
	}
	opts := bind.NewKeyedTransactor(key)
	opts.GasLimit = uint64(6721975)
	opts.GasPrice = gasPrice
	return opts, nil
}
