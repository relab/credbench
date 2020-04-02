package cmd

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"log"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/relab/bbchain-dapp/benchmark/database"
	"github.com/spf13/cobra"

	pb "github.com/relab/bbchain-dapp/benchmark/proto"
)

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "Manage accounts",
}

var createAccountsCmd = &cobra.Command{
	Use:   "create",
	Short: "Create N accounts",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		n, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatalln(err.Error())
		}

		_, err = createAccounts(n)
		if err != nil {
			log.Fatalln(err.Error())
		}
		fmt.Printf("%d accounts successfully created\n", n)
	},
}

func createAccounts(n int) ([]*pb.Account, error) {
	as, err := database.CreateAccountStore(db)
	if err != nil {
		return []*pb.Account{}, err
	}
	accounts := generateAccounts(n)
	err = as.AddAccounts([]string{"eth_accounts"}, accounts)
	if err != nil {
		return []*pb.Account{}, err
	}
	return accounts, nil
}

func generateAccounts(n int) []*pb.Account {
	accounts := make([]*pb.Account, n)
	for i := 0; i < n; i++ {
		key, address := newKey()
		hexkey := keyToHex(key)
		accounts[i] = &pb.Account{
			Address:     address.Hex(),
			HexKey:      hexkey,
			Contracts:   []string{},
			Credentials: []string{},
			Selected:    pb.Type_NONE,
		}
	}
	return accounts
}

func has0xPrefix(input string) bool {
	return len(input) >= 2 && input[0] == '0' && (input[1] == 'x' || input[1] == 'X')
}

func keyToHex(privateKey *ecdsa.PrivateKey) string {
	keyBytes := crypto.FromECDSA(privateKey)
	return hexutil.Encode(keyBytes)
}

func hexToKey(hexkey string) *ecdsa.PrivateKey {
	if has0xPrefix(hexkey) {
		hexkey = hexkey[2:]
	}
	key, err := crypto.HexToECDSA(hexkey)
	if err != nil {
		log.Panic(err)
	}
	return key
}

func newKey() (*ecdsa.PrivateKey, common.Address) {
	privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	return privateKey, address
}

func GetTxOpts(hexkey string) (*bind.TransactOpts, error) {
	backend, _ := clientConn.Backend()
	gasPrice, err := backend.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("Failed to estimate the gas price: %v", err)
	}
	opts := bind.NewKeyedTransactor(hexToKey(hexkey))
	opts.GasLimit = uint64(6721975)
	opts.GasPrice = gasPrice
	return opts, nil
}

func init() {
	rootCmd.AddCommand(accountsCmd)
	accountsCmd.AddCommand(createAccountsCmd)
}
