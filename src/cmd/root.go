package cmd

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"time"

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

var defaultWaitTime = 10 * time.Second

var rootCmd = &cobra.Command{
	Use:   "bbchain",
	Short: "BBChain verifiable credential system",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		err := setupPreRun()
		if err != nil {
			log.Fatalln(err.Error())
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err.Error())
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&datadir, "datadir", defaultDatadir(), "path to the root app directory")
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file")
	rootCmd.PersistentFlags().StringVar(&defaultAccount, "defaultAccount", "", "Ethereum default account address")
	rootCmd.PersistentFlags().StringVar(&backendURL, "backendURL", "http://127.0.0.1:8545", "Blockchain backend host:port")
	rootCmd.PersistentFlags().StringVar(&keystoreDir, "keystore", defaultKeyStore(), "Keystore root directory")
	rootCmd.PersistentFlags().StringVar(&ipcFile, "ipc",
		defaultIPC(), "Ethereum Inter-process Communication file")
	rootCmd.PersistentFlags().BoolVar(&waitPeers, "waitPeers", false, "Minimum number of peers connected")
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
}

func parseConfigFile() {
	datadir = viper.GetString("datadir")
	defaultAccount = viper.GetString("defaultAccount")
	keystoreDir = viper.GetString("keystore")
	backendURL = "http://" + viper.GetString("backend.host") + ":" + viper.GetString("backend.port")
	ipcFile = viper.GetString("backend.ipc")
	waitPeers = viper.GetBool("backend.waitPeers")
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
	return filepath.Join(currentUser.HomeDir, ".bbchain")
}

func defaultKeyStore() string {
	return filepath.Join(datadir, "/keystore")
}

func defaultIPC() string {
	return filepath.Join(datadir, "/geth.ipc")
}
