package cmd

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/relab/ct-eth-dapp/cli/transactor"
	"github.com/relab/ct-eth-dapp/src/accounts"
	"github.com/relab/ct-eth-dapp/src/ctree/aggregator"
	"github.com/relab/ct-eth-dapp/src/ctree/notary"
	"github.com/relab/ct-eth-dapp/src/deployer"
	"github.com/relab/ct-eth-dapp/src/faculty"

	keyutils "github.com/relab/ct-eth-dapp/src/accounts"
	course "github.com/relab/ct-eth-dapp/src/course"
)

func deployNotaryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "notary",
		Short: "Deploy notary library",
		Run: func(cmd *cobra.Command, args []string) {
			key, _, _ := keyutils.GetKeys(senderHexKey)
			err := deployNotary(backend, key)
			if err != nil {
				log.Fatalln(err.Error())
			}
		},
	}
}

func deployNotary(backend *ethclient.Client, key *ecdsa.PrivateKey) error {
	fmt.Println("Deploying Notary...")
	addr, tx, _, err := LinkAndDeploy(backend, key, notary.NotaryContractABI, notary.NotaryContractBin, nil)
	if err != nil {
		return fmt.Errorf("failed to deploy the contract: %v", err)
	}
	viper.Set("deployed_libs.notary", addr.Hex())
	err = viper.WriteConfig() // FIXME: this currently override the config
	if err != nil {
		return err
	}
	fmt.Printf("Transaction ID: %x\n", tx.Hash())
	return nil
}

func deployAggregatorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "aggregator",
		Short: "Deploy aggregator library",
		Run: func(cmd *cobra.Command, args []string) {
			key, _, _ := keyutils.GetKeys(senderHexKey)
			err := deployAggregator(backend, key)
			if err != nil {
				log.Fatalln(err.Error())
			}
		},
	}
}

func deployAggregator(backend *ethclient.Client, key *ecdsa.PrivateKey) error {
	fmt.Println("Deploying Aggregator...")
	addr, tx, _, err := LinkAndDeploy(backend, key, aggregator.CredentialSumABI, aggregator.CredentialSumBin, nil)
	if err != nil {
		return fmt.Errorf("failed to deploy the contract: %v", err)
	}
	viper.Set("deployed_libs.aggregator", addr.Hex())
	err = viper.WriteConfig()
	if err != nil {
		return err
	}
	fmt.Printf("Transaction ID: %x\n", tx.Hash())
	return nil
}

var deployAllLibsCmd = &cobra.Command{
	Use:   "libs",
	Short: "Deploy all libraries",
	Run: func(cmd *cobra.Command, args []string) {
		key, _, _ := keyutils.GetKeys(senderHexKey)

		err := deployNotary(backend, key)
		if err != nil {
			log.Fatalln(err.Error())
		}
		err = deployAggregator(backend, key)
		if err != nil {
			log.Fatalln(err.Error())
		}
	},
}

func deployCourseCmd() *cobra.Command {
	var owners []string
	var quorum uint8

	c := &cobra.Command{
		Use:   "course",
		Short: "Deploy course contract",
		Run: func(cmd *cobra.Command, args []string) {
			key, _, _ := keyutils.GetKeys(senderHexKey)
			var ownersAddr []common.Address
			for _, addr := range owners {
				ownersAddr = append(ownersAddr, common.HexToAddress(addr))
			}
			_, tx, err := DeployCourse(backend, key, ownersAddr, quorum)
			if err != nil {
				log.Fatalln(err.Error())
			}
			fmt.Printf("Transaction ID: %x\n", tx.Hash())
		},
	}

	c.Flags().StringSliceVar(&owners, "owners", []string{}, "Owners addresses (comma separated)")
	c.Flags().Uint8Var(&quorum, "quorum", uint8(len(owners)), "Minimum number of signatures required to issue course credentials")

	c.MarkFlagRequired("owners")
	c.MarkFlagRequired("quorum")

	return c
}

func DeployCourse(backend *ethclient.Client, key *ecdsa.PrivateKey, owners []common.Address, quorum uint8) (common.Address, *types.Transaction, error) {
	fmt.Println("Deploying Course...")
	aggregatorAddr := viper.GetString("deployed_libs.aggregator")
	if aggregatorAddr == "" {
		log.Fatalln(fmt.Errorf("Aggregator contract not deployed. Please, deploy it first"))
	}
	notaryAddr := viper.GetString("deployed_libs.notary")
	if notaryAddr == "" {
		log.Fatalln(fmt.Errorf("Notary contract not deployed. Please, deploy it first"))
	}
	libs := map[string]string{
		"CredentialSum": aggregatorAddr,
		"Notary":        notaryAddr,
	}

	cAddr, tx, _, err := LinkAndDeploy(backend, key, course.CourseContractABI, course.CourseContractBin, libs, owners, quorum)
	if err != nil {
		return common.Address{}, nil, err
	}
	if accounts.IsZeroAddress(cAddr) {
		return common.Address{}, nil, errors.New("zero address")
	}
	return cAddr, tx, nil
}

func DeployFaculty(backend *ethclient.Client, key *ecdsa.PrivateKey, owners []common.Address, quorum uint8) (common.Address, *types.Transaction, error) {
	fmt.Println("Deploying Faculty...")
	aggregatorAddr := viper.GetString("deployed_libs.aggregator")
	if aggregatorAddr == "" {
		log.Fatalln(fmt.Errorf("Aggregator contract not deployed. Please, deploy it first"))
	}
	notaryAddr := viper.GetString("deployed_libs.notary")
	if notaryAddr == "" {
		log.Fatalln(fmt.Errorf("Notary contract not deployed. Please, deploy it first"))
	}
	libs := map[string]string{
		"CredentialSum": aggregatorAddr,
		"Notary":        notaryAddr,
	}

	cAddr, tx, _, err := LinkAndDeploy(backend, key, faculty.FacultyContractABI, faculty.FacultyContractBin, libs, owners, quorum)
	if err != nil {
		return common.Address{}, nil, err
	}
	if accounts.IsZeroAddress(cAddr) {
		return common.Address{}, nil, errors.New("zero address")
	}
	return cAddr, tx, nil
}

// LinkAndDeploy links a contract with the given libraries and deploy it
// using the default account
// TODO: pass a flag to wait tx confirmations
func LinkAndDeploy(backend *ethclient.Client, key *ecdsa.PrivateKey, contractABI, contractBin string, libs map[string]string, params ...interface{}) (common.Address, *types.Transaction, *bind.BoundContract, error) {
	accAddress := accounts.GetAddress(key).Hex()

	fmt.Printf("Deployer: %s balance: %v\n", accAddress, transactor.GetBalance(accAddress, backend))

	opts, err := accounts.GetTxOpts(key, backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	if len(libs) > 0 {
		contractBin = deployer.LinkContract(contractBin, libs)
	}

	address, tx, contract, err := deployer.DeployContract(opts, backend, contractABI, contractBin, params...)
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	err = deployer.WaitForTxConfirmation(context.TODO(), backend, tx, 0)
	if err != nil {
		return address, tx, nil, fmt.Errorf("Transaction not confirmed due to error: %v", err)
	}
	fmt.Printf("Contract %s successfully deployed\n", address.Hex())
	return address, tx, contract, nil
}

func newDeployCmd() *cobra.Command {
	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy contracts",
	}

	deployCmd.AddCommand(
		deployAllLibsCmd,
		deployNotaryCmd(),
		deployAggregatorCmd(),
		deployCourseCmd(),
	)

	return deployCmd
}
