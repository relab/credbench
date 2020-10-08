package cmd

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/relab/ct-eth-dapp/src/client"
	"github.com/relab/ct-eth-dapp/src/fileutils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	configFile     string
	backendURL     string
	defaultAccount string
	keystoreDir    string
	datadir        string
	ipcFile        string
	waitPeers      bool
)

var backend *ethclient.Client

var rootCmd = &cobra.Command{
	Use:   "ctethapp",
	Short: "Ethereum Credential Transparency System",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		err := loadWallet()
		if err != nil {
			log.Fatalln(err.Error())
		}

		clientConn, err := setupClient()
		if err != nil {
			log.Fatalln(err.Error())
		}
		backend, _ = clientConn.Backend()
	},
	PersistentPostRun: func(_ *cobra.Command, _ []string) {
		backend.Close()
	},
}

func Execute() {
	rootCmd.AddCommand(
		listenCmd,
		newDeployCmd(),
		newAccountCmd(),
		newCourseCmd(),
		newBenchCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err.Error())
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&datadir, "datadir", defaultDatadir(), "path to the root app directory")
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file")
	rootCmd.PersistentFlags().StringVar(&defaultAccount, "default_account", "", "Ethereum default account address")
	rootCmd.PersistentFlags().StringVar(&backendURL, "backendURL", "http://127.0.0.1:8545", "Blockchain backend host:port")
	rootCmd.PersistentFlags().StringVar(&keystoreDir, "keystore", defaultKeyStore(), "Keystore root directory")
	rootCmd.PersistentFlags().StringVar(&ipcFile, "ipc",
		defaultIPC(), "Ethereum Inter-process Communication file")
	rootCmd.PersistentFlags().BoolVar(&waitPeers, "wait_peers", false, "Minimum number of peers connected")

	cobra.OnInitialize(initConfig)
}

func initConfig() {
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		path := defaultConfigPath()
		viper.AddConfigPath(path)
		viper.SetConfigName("config")
	}
	viper.SetConfigType("json")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
		parseConfigFile()
	}
	initSetup()
}

func initSetup() (err error) {
	err = fileutils.CreateDir(datadir)
	if err != nil {
		return err
	}
	return err
}

func setupClient() (client.EthClient, error) {
	c, err := client.NewClient(backendURL)
	if err != nil {
		return nil, err
	}

	if waitPeers {
		err = c.CheckConnectPeers(10 * time.Second)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

func parseConfigFile() {
	datadir = viper.GetString("datadir")
	defaultAccount = viper.GetString("default_account")
	keystoreDir = viper.GetString("keystore")
	backendURL = "http://" + viper.GetString("backend.host") + ":" + viper.GetString("backend.port")
	ipcFile = viper.GetString("backend.ipc")
	waitPeers = viper.GetBool("backend.wait_peers")

	fmt.Println("ROOT DATADIR: ", datadir)
}

func defaultConfigPath() string {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err.Error())
	}
	return pwd
}

func defaultDatadir() string {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf("error getting user home directory: %v", err)
	}
	return filepath.Join(currentUser.HomeDir, ".ctethapp")
}

func defaultKeyStore() string {
	return filepath.Join(datadir, "/keystore")
}

func defaultIPC() string {
	return filepath.Join(datadir, "/geth.ipc")
}
