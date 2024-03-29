package genesis

import (
	"html/template"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/relab/credbench/bench/datastore"
	pb "github.com/relab/credbench/bench/proto"
	ctaccounts "github.com/relab/credbench/pkg/accounts"
)

var (
	ChainID        = 5777
	Difficulty     = "0x1"
	GasLimit       = "6721975"
	GasPrice       = "20000000000"
	DefaultBalance = "100000000000000000000"
	CliquePeriod   = 0
	CliqueEpoch    = 2000
)

type GenesisData struct {
	ChainID        int
	Difficulty     string
	GasLimit       string
	DefaultBalance string
	N              int      // len(Accounts) - 1
	Accounts       []string // Account addresses
	ExtraData      hexutil.Bytes
	Clique         bool
	CliquePeriod   int
	CliqueEpoch    int
}

func GenerateGenesis(datadirPath string, consensus string, accountStore *datastore.EthAccountStore, n int) error {
	accounts, err := CreateAccounts(accountStore, n)
	if err != nil {
		return err
	}
	// Select one deployer (first account on the genesis)
	// We are also using it as a signer on the clique
	deployer := accounts[0].Address
	_, err = accountStore.SelectAccount(pb.Type_SEALER, deployer)
	if err != nil {
		return err
	}

	log.Infof("Configured Clique Sealer/Validator: %s\n", common.BytesToAddress(accounts[0].Address).Hex())
	genesisFile := filepath.Join(datadirPath, "genesis.json")
	return createGenesisFile(genesisFile, newGenesisData(datadirPath, consensus, accounts))
}

func newGenesisData(datadirPath string, consensus string, accounts datastore.Accounts) *GenesisData {
	if len(accounts) == 0 {
		log.Fatal("Attempt to create genesis without accounts")
		return nil
	}

	genesis := &GenesisData{
		ChainID:        ChainID,
		Difficulty:     Difficulty,
		GasLimit:       GasLimit,
		DefaultBalance: DefaultBalance,
		N:              len(accounts) - 1,
		Accounts:       accounts.ToHex(),
	}

	// TODO select n signers accounts instead of only one
	signersAccounts := datastore.Accounts{accounts[0]}
	if consensus == "clique" {
		// same test password for all accounts
		keystorePath := filepath.Join(datadirPath, "keystore")
		signers := createSignersKeystore(keystorePath, signersAccounts, "123")
		_ = createTestPasswordFile(datadirPath, "123")
		// Clique is a proof-of-authority consensus.
		// It requires extradata to be the concatenation of 32 zero bytes,
		// all signer addresses (without 0x prefix) and 65 further zero bytes.
		// https://geth.ethereum.org/docs/interface/private-network
		extraData := make([]byte, 32+len(signers)*common.AddressLength+65)
		genesis.Clique = true
		genesis.CliquePeriod = CliquePeriod
		genesis.CliqueEpoch = CliqueEpoch
		for i, signer := range signers {
			byteAddr := common.HexToAddress(signer).Bytes()
			copy(extraData[32+i*common.AddressLength:], byteAddr[:])
		}

		genesis.ExtraData = hexutil.Bytes(extraData) // default: 0x
	}
	return genesis
}

func createSignersKeystore(keystorePath string, accounts datastore.Accounts, password string) []string {
	ks := keystore.NewKeyStore(keystorePath, keystore.StandardScryptN, keystore.StandardScryptP)
	for _, account := range accounts {
		key, _, _ := ctaccounts.GetKeys(account.HexKey)
		_, _ = ks.ImportECDSA(key, password)
	}
	return accounts.ToHex()
}

func createTestPasswordFile(path string, password string) error {
	f, err := os.Create(filepath.Join(path, "password.txt"))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(password)
	if err != nil {
		return err
	}
	return nil
}

func createGenesisFile(path string, data *GenesisData) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	tmpl := template.Must(template.New("").Parse(genesisTmpl))
	if err := tmpl.Execute(f, data); err != nil {
		return err
	}

	return nil
}

func CreateAccounts(accountStore *datastore.EthAccountStore, n int) ([]*pb.Account, error) {
	accounts := generateAccounts(n)
	err := accountStore.PutAccount(accounts...)
	if err != nil {
		return []*pb.Account{}, err
	}
	log.Infof("%d accounts successfully created\n", n)
	return accounts, nil
}

func generateAccounts(n int) []*pb.Account {
	accounts := make([]*pb.Account, n)
	for i := 0; i < n; i++ {
		key, address := ctaccounts.NewKey()
		hexkey := ctaccounts.KeyToHex(key)
		accounts[i] = &pb.Account{
			Address:   hexutil.MustDecode(address.Hex()),
			HexKey:    hexkey,
			Nonce:     0,
			Contracts: [][]byte{},
			Selected:  pb.Type_NONE,
		}
	}
	return accounts
}

const genesisTmpl = `
{
    "config": {
        "chainId": {{ .ChainID }},
        "homesteadBlock": 0,
        "eip150Block": 0,
        "eip155Block": 0,
        "eip158Block": 0,
        "byzantiumBlock": 0,
        "constantinopleBlock": 0,
        "petersburgBlock": 0,
        {{if .Clique }}"clique": {
            "period": {{ .CliquePeriod }},
            "epoch": {{ .CliqueEpoch }}
        }{{ else }}"ethash": {}{{ end }}
    },
    "coinbase": "{{index .Accounts 0}}",
    "difficulty": "{{ .Difficulty }}",
    "nonce": "0x0000000000000000",
    "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
    "extradata": "{{ .ExtraData }}",
    "gasLimit": "{{ .GasLimit }}",
    "timestamp": "0x0",
    "alloc": { {{range $i, $address := .Accounts}}
        "{{ $address }}": { "balance": "{{ $.DefaultBalance }}" }{{if lt $i $.N}}, {{end}}{{ end }}
    }
}
`
