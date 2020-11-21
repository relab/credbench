package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	. "github.com/logrusorgru/aurora"
	log "github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum/common"
	"github.com/relab/ct-eth-dapp/src/ctree/node"
)

var onChain bool

var verifyCredentialTreeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Verifies a credential tree",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cAddr := common.HexToAddress(args[0])
		c, err := node.NewNode(cAddr, backend)
		if err != nil {
			log.Fatal(err)
		}
		sAddr := common.HexToAddress(args[1])
		if err = c.VerifyCredentialTree(onChain, nil, sAddr); err != nil {
			log.Fatal(err)
		}
		fmt.Println(Green("Valid"), "credential tree!")
	},
}

var verifyCredentialRootCmd = &cobra.Command{
	Use:   "root",
	Short: "Verifies a credential root",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		cAddr := common.HexToAddress(args[0])
		c, err := node.NewNode(cAddr, backend)
		if err != nil {
			log.Fatal(err)
		}
		sAddr := common.HexToAddress(args[1])
		root := common.HexToHash(args[2])
		if err = c.VerifyCredentialRoot(onChain, nil, sAddr, root); err != nil {
			log.Fatal(err)
		}
		fmt.Println(Green("Valid"), "credential root!")
	},
}

var verifyCredentialCmd = &cobra.Command{
	Use:   "credential",
	Short: "Verifies a credential",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		cAddr := common.HexToAddress(args[0])
		sAddr := common.HexToAddress(args[1])
		digest := common.HexToHash(args[2])
		c, err := node.NewNode(cAddr, backend)
		if err != nil {
			log.Fatal(err)
		}
		if err := c.VerifyCredential(onChain, nil, sAddr, digest); err != nil {
			log.Fatal(err)
		}
		fmt.Println(Green("Valid"), "credential!")
	},
}

var verifyIssuedCredentialsCmd = &cobra.Command{
	Use:   "all",
	Short: "Verifies all issued credentials of the given contract for the given subject",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cAddr := common.HexToAddress(args[0])
		c, err := node.NewNode(cAddr, backend)
		if err != nil {
			log.Fatal(err)
		}
		sAddr := common.HexToAddress(args[1])
		if err := c.VerifyIssuedCredentials(onChain, nil, sAddr); err != nil {
			log.Fatal(err)
		}
		fmt.Println("All credentials are", Green("valid"))
	},
}

func newVerifyCmd() *cobra.Command {
	verifyCmd := &cobra.Command{
		Use:   "verify",
		Short: "Verifies a credential",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			rootCmd.PersistentPreRun(cmd, args)
			err := loadDefaultAccount()
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	verifyCmd.PersistentFlags().BoolVar(&onChain, "onchain", false, "perform an on-chain/off-chain verification")

	verifyCmd.AddCommand(verifyCredentialCmd)
	verifyCmd.AddCommand(verifyIssuedCredentialsCmd)
	verifyCmd.AddCommand(verifyCredentialTreeCmd)
	verifyCmd.AddCommand(verifyCredentialRootCmd)
	return verifyCmd
}
