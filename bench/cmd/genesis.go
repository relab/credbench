package cmd

import (
	"errors"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/relab/ct-eth-dapp/bench/datastore"
	"github.com/relab/ct-eth-dapp/bench/genesis"
	"github.com/relab/ct-eth-dapp/bench/helm"
	pb "github.com/relab/ct-eth-dapp/bench/proto"
	keyutils "github.com/relab/ct-eth-dapp/pkg/accounts"
)

var genesisCmd = &cobra.Command{
	Use:   "genesis",
	Short: "Generate genesis file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		n, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatal(err)
		}
		err = genesis.GenerateGenesis(datadir, consensus, accountStore, n)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func getAccountAddresses() ([]string, error) {
	accounts, err := accountStore.All()
	if err != nil {
		return nil, err
	}
	return accounts.ToHex(), nil
}

var exportHelmCmd = &cobra.Command{
	Use:   "helm",
	Short: "Export helm file",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("please specify the number of replicas")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		replicaCount, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatal(err)
		}

		addresses, err := getAccountAddresses()
		if err != nil {
			log.Fatal(err)
		}

		// Get initial validator account
		validators, err := accountStore.GetByType(1, pb.Type_SEALER)
		if err != nil {
			log.Fatal(err)
		}
		validatorAccount := validators[0]
		// address and hexkey should be without the 0x prefix
		validatorAddress := datastore.GetStringAddress(validatorAccount)
		validatorKey := validatorAccount.GetHexKey()
		if keyutils.Has0xPrefix(validatorKey) {
			validatorKey = validatorKey[2:]
		}
		err = helm.ExportHelmFile(datadir, genesis.ChainID, genesis.GasLimit, genesis.GasPrice, genesis.DefaultBalance, validatorAddress, validatorKey, replicaCount, addresses)
		if err != nil {
			log.Fatal(err)
		}
	},
}
