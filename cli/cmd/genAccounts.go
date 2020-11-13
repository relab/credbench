package cmd

import (
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/relab/ct-eth-dapp/cli/genesis"
	"github.com/spf13/cobra"
)

var createAccountsCmd = &cobra.Command{
	Use:   "create",
	Short: "Create N accounts",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		n, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatal(err)
		}

		_, err = genesis.CreateAccounts(accountStore, n)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func newGenAccountsCmd() *cobra.Command {
	genAccountsCmd := &cobra.Command{
		Use:   "accounts",
		Short: "Generate accounts",
	}

	genAccountsCmd.AddCommand(createAccountsCmd)
	return genAccountsCmd
}
