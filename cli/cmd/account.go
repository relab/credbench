package cmd

import (
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	"github.com/relab/ct-eth-dapp/cli/transactor"
	"github.com/relab/ct-eth-dapp/src/accounts"
)

var (
	keyStore *keystore.KeyStore
	wallet   accounts.Wallet
)

func loadWallet() (err error) {
	senderAddr := accounts.GetAccountAddress(defaultAccount, keystoreDir)
	wallet, err = accounts.NewWallet(senderAddr, keystoreDir)
	return err
}

func loadKeystore() {
	keyStore = keystore.NewKeyStore(keystoreDir, keystore.StandardScryptN, keystore.StandardScryptP)
}

var createAccountCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new account",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		err := accounts.NewAccount(keystoreDir)
		if err != nil {
			log.Fatalln(err.Error())
		}
	},
}

var importAccountCmd = &cobra.Command{
	Use:   "import",
	Short: "Import an account based on private key",
	PreRun: func(cmd *cobra.Command, args []string) {
		loadKeystore()
	},
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		wallet, err = accounts.ImportKey(args[0], keyStore)
		if err != nil {
			log.Fatalln(err.Error())
		}
	},
}

var getBalanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Get balance of an account",
	PreRun: func(cmd *cobra.Command, args []string) {
		loadKeystore()
	},
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		address := common.HexToAddress(args[0])
		balance := transactor.GetBalance(address, backend)
		fmt.Printf("Balance of account %s: %v\n", args[0], balance)
	},
}

func newAccountCmd() *cobra.Command {
	accountCmd := &cobra.Command{
		Use:   "account",
		Short: "Manage accounts",
	}
	accountCmd.AddCommand(
		createAccountCmd,
		importAccountCmd,
		getBalanceCmd,
	)
	return accountCmd
}
