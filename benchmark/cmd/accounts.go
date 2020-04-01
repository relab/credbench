package cmd

import (
	"log"
	"strconv"

	"github.com/relab/bbchain-dapp/benchmark/database"
	"github.com/spf13/cobra"

	pb "github.com/relab/bbchain-dapp/benchmark/proto"
)

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "Manage accounts",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		err := setupDB(dbPath, dbFile)
		if err != nil {
			log.Fatalln(err.Error())
		}
	},
	PersistentPostRun: func(_ *cobra.Command, _ []string) {
		db.Close()
	},
}

var createAccountsCmd = &cobra.Command{
	Use:   "create",
	Short: "Create N accounts",
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
func init() {
	rootCmd.AddCommand(accountsCmd)
	accountsCmd.AddCommand(createAccountsCmd)
}
