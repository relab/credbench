package cmd

import (
	"log"

	"github.com/relab/bbchain-dapp/src/core/accounts"
	"github.com/spf13/cobra"
)

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Manage accounts",
}

var newAccount = &cobra.Command{
	Use:   "new",
	Short: "Create a new account",
	Run: func(cmd *cobra.Command, args []string) {
		err := accounts.NewAccount(keystoreDir)
		if err != nil {
			log.Fatalln(err.Error())
		}
	},
}

var importAccount = &cobra.Command{
	Use:   "import",
	Short: "Import an account based on private key",
	PreRun: func(cmd *cobra.Command, args []string) {
		loadKeystore()
	},
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		wallet, err = accounts.ImportKey(args[0], keyStore)
		if err != nil {
			log.Fatalln(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(accountCmd)
	accountCmd.AddCommand(newAccount)
	accountCmd.AddCommand(importAccount)
}
