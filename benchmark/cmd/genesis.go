package cmd

import (
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/relab/ct-eth-dapp/benchmark/datastore"
	keyutils "github.com/relab/ct-eth-dapp/src/core/accounts"
)

var (
	chainID        = 42
	difficulty     = "1"
	gasLimit       = "12460000"
	defaultBalance = "10000000000000000000"
)

type GenesisData struct {
	ChainID        int
	Difficulty     string
	GasLimit       string
	DefaultBalance string
	N              int      // len(Accounts) - 1
	Accounts       []string // Account addresses
	ExtraData      hexutil.Bytes
	POA            bool
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
	return createGenesisFile(newGenesisData(accounts))
}

func newGenesisData(accounts datastore.Accounts) *GenesisData {
	if len(accounts) == 0 {
		log.Fatalln("Attempt to create genesis without accounts")
		return nil
	}

	genesis := &GenesisData{
		ChainID:        chainID,
		Difficulty:     difficulty,
		GasLimit:       gasLimit,
		DefaultBalance: defaultBalance,
		N:              len(accounts) - 1,
		Accounts:       accounts.ToHex(),
	}

	//TODO select n signers accounts
	signersAccounts := datastore.Accounts{accounts[0]}
	if consensus == "poa" {
		// same test password for all accounts
		signers := createSignersKeystore(signersAccounts, "123")
		createTestPasswordFile("123")
		// POA requires extradata to be the concatenation of 32 zero bytes,
		// all signer addresses (without 0x prefix) and 65 further zero bytes.
		// https://geth.ethereum.org/docs/interface/private-network
		extraData := make([]byte, 32+len(signers)*common.AddressLength+65)
		genesis.POA = true
		for i, signer := range signers {
			byteAddr := common.HexToAddress(signer).Bytes()
			copy(extraData[32+i*common.AddressLength:], byteAddr[:])
		}
		genesis.ExtraData = hexutil.Bytes(extraData) //default: 0x
	}
	return genesis
}

func createSignersKeystore(accounts datastore.Accounts, password string) []string {
	ks := keystore.NewKeyStore(keystoreDir, keystore.StandardScryptN, keystore.StandardScryptP)
	for _, account := range accounts {
		key, _, _ := keyutils.GetKeys(account.HexKey)
		ks.ImportECDSA(key, password)
	}
	return accounts.ToHex()
}

func createTestPasswordFile(password string) error {
	f, err := os.Create(filepath.Join(datadir, "password.txt"))
	defer f.Close()

	_, err = f.WriteString(password)
	if err != nil {
		return err
	}
	return nil
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
