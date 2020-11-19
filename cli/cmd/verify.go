package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	. "github.com/logrusorgru/aurora"
	log "github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum/common"
	"github.com/relab/ct-eth-dapp/src/ctree"
)

var (
	onChain bool
	cType   string
)

var verifyCredentialTreeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Verifies a credential tree",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cAddr := common.HexToAddress(args[0])
		c, err := getContract(cAddr)
		if err != nil {
			log.Fatal(err)
		}
		sAddr := common.HexToAddress(args[1])
		ok, err := c.VerifyCredentialTree(onChain, nil, sAddr)
		if err != nil {
			log.Fatal(err)
		}
		if ok {
			fmt.Println(Green("Valid"), "credential tree!")
		} else {
			fmt.Println(Red("Invalid"), "credential tree!")
		}
	},
}

var verifyCredentialRootCmd = &cobra.Command{
	Use:   "root",
	Short: "Verifies a credential root",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		cAddr := common.HexToAddress(args[0])
		c, err := getContract(cAddr)
		if err != nil {
			log.Fatal(err)
		}
		sAddr := common.HexToAddress(args[1])
		root := common.HexToHash(args[2])
		ok, err := c.VerifyCredentialRoot(onChain, nil, sAddr, root)
		if err != nil {
			log.Fatal(err)
		}
		if ok {
			fmt.Println(Green("Valid"), "credential root!")
		} else {
			fmt.Println(Red("Invalid"), "credential root!")
		}
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
		c, err := getContract(cAddr)
		if err != nil {
			log.Fatal(err)
		}
		ok, err := c.VerifyCredential(onChain, nil, sAddr, digest)
		if err != nil {
			log.Fatal(err)
		}
		if ok {
			fmt.Println(Green("Valid"), "credential!")
		} else {
			fmt.Println(Red("Invalid"), "credential!")
		}
	},
}

var verifyIssuedCredentialsCmd = &cobra.Command{
	Use:   "all",
	Short: "Verifies all issued credentials of the given contract for the given subject",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cAddr := common.HexToAddress(args[0])
		c, err := getContract(cAddr)
		if err != nil {
			log.Fatal(err)
		}
		sAddr := common.HexToAddress(args[1])
		ok, err := c.VerifyIssuedCredentials(onChain, nil, sAddr)
		if err != nil {
			log.Fatal(err)
		}
		if ok {
			fmt.Println("All credentials are", Green("valid"))
		} else {
			fmt.Println("There are", Red("invalid"), "credentials")
		}
	},
}

func getContract(address common.Address) (c ctree.Verifier, err error) {
	switch cType {
	case "course":
		c, err = getCourseContract(address)
		if err != nil {
			return nil, err
		}
	case "faculty":
		c, err = getFacultyContract(address)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("wrong contract type")
	}
	return c, err
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
	verifyCmd.PersistentFlags().StringVar(&cType, "contract", "course", "contract type (course/faculty)")

	verifyCmd.AddCommand(verifyCredentialCmd)
	verifyCmd.AddCommand(verifyIssuedCredentialsCmd)
	verifyCmd.AddCommand(verifyCredentialTreeCmd)
	verifyCmd.AddCommand(verifyCredentialRootCmd)
	return verifyCmd
}
