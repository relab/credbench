package cmd

import (
	"errors"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/relab/ct-eth-dapp/src/core/accounts"
	course "github.com/relab/ct-eth-dapp/src/core/course"
	"github.com/relab/ct-eth-dapp/src/ctree/aggregator"
	"github.com/relab/ct-eth-dapp/src/ctree/notary"
	"github.com/relab/ct-eth-dapp/src/transactor"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy contracts",
}

func deployNotaryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "notary",
		Short: "Deploy notary library",
		Run: func(cmd *cobra.Command, args []string) {
			err := deployNotary()
			if err != nil {
				log.Fatalln(err.Error())
			}
		},
	}
}

func deployNotary() error {
	addr, tx, _, err := LinkAndDeploy(notary.NotaryContractABI, notary.NotaryContractBin, nil)
	if err != nil {
		return fmt.Errorf("failed to deploy the contract: %v", err)
	}
	viper.Set("deployed_libs.notary", addr.Hex())
	viper.WriteConfig() //FIXME: this currently override the config
	fmt.Printf("Transaction ID: %x\n", tx.Hash())
	return nil
}

func deployAggregatorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "aggregator",
		Short: "Deploy aggregator library",
		Run: func(cmd *cobra.Command, args []string) {
			err := deployAggregator()
			if err != nil {
				log.Fatalln(err.Error())
			}
		},
	}
}

func deployAggregator() error {
	addr, tx, _, err := LinkAndDeploy(aggregator.CredentialSumABI, aggregator.CredentialSumBin, nil)
	if err != nil {
		return fmt.Errorf("failed to deploy the contract: %v", err)
	}
	viper.Set("deployed_libs.aggregator", addr.Hex())
	viper.WriteConfig()
	fmt.Printf("Transaction ID: %x\n", tx.Hash())
	return nil
}

var deployAllLibsCmd = &cobra.Command{
	Use:   "libs",
	Short: "Deploy all libraries",
	Run: func(cmd *cobra.Command, args []string) {
		err := deployNotary()
		if err != nil {
			log.Fatalln(err.Error())
		}
		err = deployAggregator()
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
			tx, err := deployCourse(owners, quorum)
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

func deployCourse(ownersList []string, quorum uint8) (*types.Transaction, error) {
	var owners []common.Address
	for _, addr := range ownersList {
		owners = append(owners, common.HexToAddress(addr))
	}

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

	cAddr, tx, _, err := LinkAndDeploy(course.CourseContractABI, course.CourseContractBin, libs, owners, quorum)
	if err != nil {
		return nil, err
	}
	if accounts.IsZeroAddress(cAddr) {
		return nil, errors.New("zero address")
	}
	return tx, nil
}

// LinkAndDeploy links a contract with the given libraries and deploy it
// using the default account
func LinkAndDeploy(contractABI, contractBin string, libs map[string]string, params ...interface{}) (common.Address, *types.Transaction, *bind.BoundContract, error) {
	backend, _ := clientConn.Backend()

	fmt.Printf("Deployer: %s balance: %v\n", wallet.HexAddress(), transactor.GetBalance(wallet.HexAddress(), backend))

	opts, err := accounts.GetTxOpts(wallet.PrivateKey(), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	if len(libs) > 0 {
		contractBin = transactor.LinkContract(contractBin, libs)
	}

	addr, tx, contract, err := transactor.DeployContract(opts, backend, contractABI, contractBin, params...)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	fmt.Printf("Contract %s successfully deployed\n", addr.Hex())

	fmt.Printf("Deployer: %s balance: %v\n", wallet.HexAddress(), transactor.GetBalance(wallet.HexAddress(), backend))
	return addr, tx, contract, nil
}

func init() {
	deployCmd.AddCommand(deployAllLibsCmd)
	deployCmd.AddCommand(deployNotaryCmd())
	deployCmd.AddCommand(deployAggregatorCmd())
	deployCmd.AddCommand(deployCourseCmd())
	rootCmd.AddCommand(deployCmd)
}
