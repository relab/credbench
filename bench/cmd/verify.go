package cmd

import (
	"fmt"
	"time"

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
		start := time.Now()
		if err = c.VerifyCredentialTree(onChain, nil, sAddr); err != nil {
			elapsed := time.Since(start)
			log.Fatalf("Verification failed in %v with error: %v\n", elapsed, err)
		}
		elapsed := time.Since(start)
		fmt.Printf("%s credential tree! Verified in %v\n", Green("Valid"), elapsed)
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
		start := time.Now()
		if err = c.VerifyCredentialRoot(onChain, nil, sAddr, root); err != nil {
			elapsed := time.Since(start)
			log.Fatalf("Verification failed in %v with error: %v\n", elapsed, err)
		}
		elapsed := time.Since(start)
		fmt.Printf("%s credential root! Verified in %v\n", Green("Valid"), elapsed)
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
		start := time.Now()
		if err := c.VerifyCredential(onChain, nil, sAddr, digest); err != nil {
			elapsed := time.Since(start)
			log.Fatalf("Verification failed in %v with error: %v\n", elapsed, err)
		}
		elapsed := time.Since(start)
		fmt.Printf("%s credential! Verified in %v\n", Green("Valid"), elapsed)
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
		start := time.Now()
		if err := c.VerifyIssuedCredentials(onChain, nil, sAddr); err != nil {
			elapsed := time.Since(start)
			log.Fatalf("Verification failed in %v with error: %v\n", elapsed, err)
		}
		elapsed := time.Since(start)
		fmt.Printf("All credentials are %s! Verified in %v\n", Green("Valid"), elapsed)
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

	verifyCmd.AddCommand(
		verifyCredentialCmd,
		verifyIssuedCredentialsCmd,
		verifyCredentialTreeCmd,
		verifyCredentialRootCmd,
	)
	return verifyCmd
}
