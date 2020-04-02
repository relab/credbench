package cmd

import (
	"html/template"
	"log"
	"os"
	"strconv"

	"github.com/relab/bbchain-dapp/benchmark/database"
	"github.com/spf13/cobra"
)

type genesisData struct {
	ChainID        int
	GasLimit       string
	DefaultBalance string
	N              int      // len(Accounts) - 1
	Accounts       []string // Account addresses
}

var genesisCmd = &cobra.Command{
	Use:   "genesis",
	Short: "Generate genesis file",
	Run: func(cmd *cobra.Command, args []string) {
		n, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatalln(err.Error())
		}

		accounts, err := createAccounts(n)
		if err != nil {
			log.Fatalln(err.Error())
		}

		addresses := database.HexAddresses(accounts)
		data := &genesisData{
			ChainID:        42,
			GasLimit:       "6721975",
			DefaultBalance: "10000000000000000000",
			N:              len(addresses) - 1,
			Accounts:       addresses,
		}

		createGenesis(data)
	},
}

func createGenesis(data *genesisData) error {
	f, err := os.Create(genesisFile)
	if err != nil {
		return err
	}

	tmpl := template.Must(template.ParseFiles(genesisTemplateFile))
	if err := tmpl.Execute(f, data); err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(genesisCmd)
}
