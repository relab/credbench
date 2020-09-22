package cmd

import (
	"html/template"
	"log"
	"os"
	"strconv"

	"github.com/relab/ct-eth-dapp/benchmark/datastore"
	"github.com/spf13/cobra"
)

type GenesisData struct {
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
		err = generateGenesis(n)
		if err != nil {
			log.Fatalln(err.Error())
		}
	},
}

func generateGenesis(n int) error {
	accounts, err := createAccounts(n)
	if err != nil {
		return err
	}
	return createGenesisFile(NewGenesisData(accounts))
}

func NewGenesisData(accounts datastore.Accounts) *GenesisData {
	if len(accounts) == 0 {
		log.Fatalln("Attempt to create genesis without accounts")
		return nil
	}
	return &GenesisData{
		ChainID:        42,
		GasLimit:       "6721975",
		DefaultBalance: "10000000000000000000",
		N:              len(accounts) - 1,
		Accounts:       accounts.ToHex(),
	}
}

func createGenesisFile(data *GenesisData) error {
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
