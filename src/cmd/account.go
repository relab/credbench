package cmd

import (
	"github.com/r0qs/dvcp/src/core/accounts"
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
		accounts.NewAccount(accountAddress)
	},
}

var importAccount = &cobra.Command{
	Use:    "import",
	Short:  "Import an account based on private key",
	PreRun: loadKeystore,
	Run: func(cmd *cobra.Command, args []string) {
		hexKey := args[0]
		wallet, _ = accounts.ImportKey(hexKey, keyStore)
	},
}

func init() {
	rootCmd.AddCommand(accountCmd)
	accountCmd.AddCommand(newAccount)
	accountCmd.AddCommand(importAccount)
}
