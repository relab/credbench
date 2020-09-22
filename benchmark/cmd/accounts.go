package cmd

import (
	"log"
	"strconv"

	"github.com/ethereum/go-ethereum/common/hexutil"
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
	},
}

func createAccounts(n int) ([]*pb.Account, error) {
	accounts := generateAccounts(n)
	err := accountStore.PutAccount(accounts...)
	if err != nil {
		return []*pb.Account{}, err
	}
	log.Printf("%d accounts successfully created\n", n)
	return accounts, nil
}

func generateAccounts(n int) []*pb.Account {
	accounts := make([]*pb.Account, n)
	for i := 0; i < n; i++ {
		key, address := keyutils.NewKey()
		hexkey := keyutils.KeyToHex(key)
		accounts[i] = &pb.Account{
			HexAddress:        hexutil.Encode(address.Bytes()),
			HexKey:            hexkey,
			ContractAddresses: []string{},
			CredentialDigests: [][]byte{},
			Selected:          pb.Type_NONE,
		}
	}
	return accounts
}

func init() {
	rootCmd.AddCommand(accountsCmd)
	accountsCmd.AddCommand(createAccountsCmd)
}
