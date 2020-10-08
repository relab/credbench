package cmd

import (
	"log"
	"strconv"

	"github.com/relab/ct-eth-dapp/cli/genesis"
	"github.com/spf13/cobra"
)

var genesisCmd = &cobra.Command{
	Use:   "genesis",
	Short: "Generate genesis file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		n, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatalln(err.Error())
		}
		err = genesis.GenerateGenesis(datadir, consensus, accountStore, n)
		if err != nil {
			log.Fatalln(err.Error())
		}
	},
}
