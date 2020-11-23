package cmd

import (
	"context"
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/relab/ct-eth-dapp/bench/transactor"
	"github.com/relab/ct-eth-dapp/src/accounts"
	"github.com/relab/ct-eth-dapp/src/ctree/aggregator"
	"github.com/relab/ct-eth-dapp/src/ctree/notary"
	"github.com/relab/ct-eth-dapp/src/deployer"
	"github.com/relab/ct-eth-dapp/src/faculty"

	course "github.com/relab/ct-eth-dapp/src/course"
)

func deployNotaryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "notary",
		Short: "Deploy notary library",
		Run: func(cmd *cobra.Command, args []string) {
			opts, err := accountStore.GetTxOpts(defaultSender.Bytes(), backend)
			if err != nil {
				log.Fatal(err)
			}

			err = deployNotary(opts, backend)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
}

func deployNotary(opts *bind.TransactOpts, backend *ethclient.Client) error {
	log.Infoln("Deploying Notary...")
	addr, tx, _, err := LinkAndDeploy(opts, backend, notary.NotaryContractABI, notary.NotaryContractBin, nil, true)
	if err != nil {
		return fmt.Errorf("failed to deploy the contract: %v", err)
	}
	viper.Set("deployed_libs.notary", addr.Hex())
	err = viper.WriteConfig() // FIXME: this currently override the config
	if err != nil {
		return err
	}
	log.Infof("Transaction ID: %x\n", tx.Hash())
	return nil
}

func deployAggregatorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "aggregator",
		Short: "Deploy aggregator library",
		Run: func(cmd *cobra.Command, args []string) {
			opts, err := accountStore.GetTxOpts(defaultSender.Bytes(), backend)
			if err != nil {
				log.Fatal(err)
			}

			err = deployAggregator(opts, backend)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
}

func deployAggregator(opts *bind.TransactOpts, backend *ethclient.Client) error {
	log.Infoln("Deploying Aggregator...")
	addr, tx, _, err := LinkAndDeploy(opts, backend, aggregator.CredentialSumABI, aggregator.CredentialSumBin, nil, true)
	if err != nil {
		return fmt.Errorf("failed to deploy the contract: %v", err)
	}
	viper.Set("deployed_libs.aggregator", addr.Hex())
	err = viper.WriteConfig()
	if err != nil {
		return err
	}
	log.Infof("Transaction ID: %x\n", tx.Hash())
	return nil
}

var deployAllLibsCmd = &cobra.Command{
	Use:   "libs",
	Short: "Deploy all libraries",
	Run: func(cmd *cobra.Command, args []string) {
		opts, err := accountStore.GetTxOpts(defaultSender.Bytes(), backend)
		if err != nil {
			log.Fatal(err)
		}

		err = deployNotary(opts, backend)
		if err != nil {
			log.Fatal(err)
		}

		opts, err = accountStore.GetTxOpts(defaultSender.Bytes(), backend)
		if err != nil {
			log.Fatal(err)
		}
		err = deployAggregator(opts, backend)
		if err != nil {
			log.Fatal(err)
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
			var ownersAddr []common.Address
			for _, addr := range owners {
				ownersAddr = append(ownersAddr, common.HexToAddress(addr))
			}

			opts, err := accountStore.GetTxOpts(defaultSender.Bytes(), backend)
			if err != nil {
				log.Fatal(err)
			}

			_, tx, err := DeployCourse(opts, backend, ownersAddr, quorum)
			if err != nil {
				log.Fatal(err)
			}
			log.Infof("Transaction ID: %x\n", tx.Hash())
		},
	}

	c.Flags().StringSliceVar(&owners, "owners", []string{}, "Owners addresses (comma separated)")
	c.Flags().Uint8Var(&quorum, "quorum", uint8(len(owners)), "Minimum number of signatures required to issue course credentials")

	c.MarkFlagRequired("owners")
	c.MarkFlagRequired("quorum")

	return c
}

func DeployCourse(opts *bind.TransactOpts, backend *ethclient.Client, owners []common.Address, quorum uint8) (common.Address, *types.Transaction, error) {
	log.Infoln("Deploying Course...")
	aggregatorAddr := viper.GetString("deployed_libs.aggregator")
	if aggregatorAddr == "" {
		log.Fatal(fmt.Errorf("Aggregator contract not deployed. Please, deploy it first"))
	}
	notaryAddr := viper.GetString("deployed_libs.notary")
	if notaryAddr == "" {
		log.Fatal(fmt.Errorf("Notary contract not deployed. Please, deploy it first"))
	}
	libs := map[string]string{
		"CredentialSum": aggregatorAddr,
		"Notary":        notaryAddr,
	}

	cAddr, tx, _, err := LinkAndDeploy(opts, backend, course.CourseContractABI, course.CourseContractBin, libs, false, owners, quorum)
	if err != nil {
		return common.Address{}, nil, err
	}
	if accounts.IsZeroAddress(cAddr) {
		return common.Address{}, nil, errors.New("zero address")
	}
	return cAddr, tx, nil
}

func DeployFaculty(opts *bind.TransactOpts, backend *ethclient.Client, owners []common.Address, quorum uint8) (common.Address, *types.Transaction, error) {
	log.Infoln("Deploying Faculty...")
	aggregatorAddr := viper.GetString("deployed_libs.aggregator")
	if aggregatorAddr == "" {
		log.Fatal(fmt.Errorf("Aggregator contract not deployed. Please, deploy it first"))
	}
	notaryAddr := viper.GetString("deployed_libs.notary")
	if notaryAddr == "" {
		log.Fatal(fmt.Errorf("Notary contract not deployed. Please, deploy it first"))
	}
	libs := map[string]string{
		"CredentialSum": aggregatorAddr,
		"Notary":        notaryAddr,
	}

	cAddr, tx, _, err := LinkAndDeploy(opts, backend, faculty.FacultyContractABI, faculty.FacultyContractBin, libs, false, owners, quorum)
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
func LinkAndDeploy(opts *bind.TransactOpts, backend *ethclient.Client, contractABI, contractBin string, libs map[string]string, waitConfirmation bool, params ...interface{}) (common.Address, *types.Transaction, *bind.BoundContract, error) {
	log.Infof("Deployer: %s balance: %v\n", opts.From.Hex(), transactor.GetBalance(opts.From, backend))

	if len(libs) > 0 {
		contractBin = deployer.LinkContract(contractBin, libs)
	}

	address, tx, contract, err := deployer.DeployContract(opts, backend, contractABI, contractBin, params...)
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	if waitConfirmation {
		err = deployer.WaitTxConfirmation(context.TODO(), backend, tx, 0)
		if err != nil {
			return address, tx, nil, fmt.Errorf("Transaction not confirmed due to error: %v", err)
		}
	}
	log.Infof("Contract %s successfully deployed\n", address.Hex())
	return address, tx, contract, nil
}

func newDeployCmd() *cobra.Command {
	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy contracts",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			rootCmd.PersistentPreRun(cmd, args)
			err := loadDefaultAccount()
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	deployCmd.AddCommand(
		deployAllLibsCmd,
		deployNotaryCmd(),
		deployAggregatorCmd(),
		deployCourseCmd(),
	)

	return deployCmd
}
