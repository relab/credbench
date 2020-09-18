package cmd

import (
	"fmt"
	"log"
	"strconv"

	"github.com/spf13/cobra"

	pb "github.com/relab/ct-eth-dapp/benchmark/proto"
	keyutils "github.com/relab/ct-eth-dapp/src/core/accounts"
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
	accounts := generateAccounts(n)
	err := accountStore.AddAccounts(accounts)
	if err != nil {
		return []*pb.Account{}, err
	}
	return accounts, nil
}

func generateAccounts(n int) []*pb.Account {
	accounts := make([]*pb.Account, n)
	for i := 0; i < n; i++ {
		key, address := keyutils.NewKey()
		hexkey := keyutils.KeyToHex(key)
		accounts[i] = &pb.Account{
			Address:     address.Bytes(),
			HexKey:      hexkey,
			Contracts:   [][]byte{},
			Credentials: [][]byte{},
			Selected:    pb.Type_NONE,
		}
	}
	return accounts
}

func init() {
	rootCmd.AddCommand(accountsCmd)
	accountsCmd.AddCommand(createAccountsCmd)
}
