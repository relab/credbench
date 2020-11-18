package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"

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
			log.Fatal(err)
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
			log.Fatal(err)
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
		log.Infof("Balance of account %s: %v\n", args[0], balance)
	},
}

var getAccountCmd = &cobra.Command{
	Use:   "get",
	Short: "Shows the account details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		address := common.HexToAddress(args[0])
		account, err := accountStore.GetAccount(address.Bytes())
		if err != nil {
			log.Error(err)
		}
		fmt.Printf("Account Info:\n")
		fmt.Printf("\tAddress: %s\n", address.Hex())
		fmt.Printf("\tHexKey: %s\n", account.HexKey)
		fmt.Printf("\tType: %v\n", account.Selected)
		fmt.Printf("\tNonce: %v\n", account.Nonce)
		fmt.Printf("\tContracts:\n")
		for _, c := range account.Contracts {
			fmt.Printf("\t  %s\n", common.BytesToAddress(c).Hex())
		}
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
		getAccountCmd,
	)
	return accountCmd
}
