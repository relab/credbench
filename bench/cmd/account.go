package cmd

import (
	"crypto/ecdsa"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"

	"github.com/relab/ct-eth-dapp/bench/eth"
	"github.com/relab/ct-eth-dapp/bench/genesis"
	pb "github.com/relab/ct-eth-dapp/bench/proto"
)

var createAccountCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new account",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		a, err := genesis.CreateAccounts(accountStore, 1)
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("Account %s successfully created", common.BytesToAddress(a[0].Address))
	},
}

func getAddress(hexkey string) common.Address {
	pkey, err := crypto.HexToECDSA(hexkey)
	if err != nil {
		log.Fatal(err)
	}
	publicKeyECDSA, ok := pkey.Public().(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}
	return crypto.PubkeyToAddress(*publicKeyECDSA)
}

func addAccount(address common.Address, hexkey string) error {
	account := &pb.Account{
		Address:   address.Bytes(),
		HexKey:    hexkey,
		Nonce:     0,
		Contracts: [][]byte{},
		Selected:  pb.Type_NONE,
	}
	err := accountStore.PutAccount(account)
	if err != nil {
		return err
	}
	return nil
}

var importAccountCmd = &cobra.Command{
	Use:   "import",
	Short: "Import an account based on hex private key",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		hexkey := args[0]
		err = addAccount(getAddress(hexkey), hexkey)
		if err != nil {
			log.Fatal(err)
		}
	},
}

var getBalanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Get balance of an account",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		address := common.HexToAddress(args[0])
		balance, err := eth.GetBalance(address, backend)
		if err != nil {
			log.Error(err)
		}
		fmt.Printf("Balance of account %s: %v\n", args[0], balance)
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
		balance, err := eth.GetBalance(address, backend)
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
		fmt.Printf("\tBalance: %v (ether)\n", balance)
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
