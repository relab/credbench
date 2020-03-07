package cmd

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	configFile     string
	backendURL     string
	accountAddress string
	keystoreDir    string
)

var rootCmd = &cobra.Command{
	Use:   "bbchain",
	Short: "BBChain verifiable credential system",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is $PWD/dev-config.json)")
	rootCmd.PersistentFlags().StringVar(&accountAddress, "accountAddress", "", "Ethereum account address")
	rootCmd.PersistentFlags().StringVar(&backendURL, "backendURL", "http://127.0.0.1:8545", "Blockchain backend host:port")
	rootCmd.PersistentFlags().StringVar(&keystoreDir, "keystore", defaultKeyStore(), "Keystore root directory")
}

func defaultKeyStore() string {
	currentUser, err := user.Current()
	if err != nil {
		fmt.Printf("error getting user home directory: %v", err)
		os.Exit(1)
	}
	return filepath.Join(currentUser.HomeDir, "/bbchain_keystore")
}
