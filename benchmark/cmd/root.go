package cmd

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"time"

	"github.com/relab/bbchain-dapp/benchmark/database"
	"github.com/relab/bbchain-dapp/src/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	configFile          string
	backendURL          string
	defaultAccount      string
	genesisFile         string
	genesisTemplateFile string
	datadir             string
	ipcFile             string
	dbPath              string
	dbFile              string
	waitPeers           bool
)

var defaultWaitTime = 10 * time.Second

var rootCmd = &cobra.Command{
	Use:   "bbchain",
	Short: "BBChain verifiable credential system",
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
	rootCmd.PersistentFlags().StringVar(&ipcFile, "ipc",
		defaultIPC(), "Ethereum Inter-process Communication file")
	rootCmd.PersistentFlags().BoolVar(&waitPeers, "waitPeers", false, "Minimum number of peers connected")
	rootCmd.PersistentFlags().StringVar(&dbPath, "dbPath", "", "Path to the database file")
	rootCmd.PersistentFlags().StringVar(&dbFile, "dbFile", "", "File name of the database")
	rootCmd.PersistentFlags().StringVar(&genesisFile, "genesisFile", "", "Path to the ethereum genesis file")

	//FIXME: should be relative to code path
	pwd, _ := os.Getwd()
	genesisTemplateFile = path.Join(pwd, "genesis/genesis.tmpl")
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
	createDatabase()
}

func parseConfigFile() {
	datadir = viper.GetString("datadir")
	defaultAccount = viper.GetString("defaultAccount")
	backendURL = "http://" + viper.GetString("backend.host") + ":" + viper.GetString("backend.port")
	ipcFile = viper.GetString("backend.ipc")
	waitPeers = viper.GetBool("backend.waitPeers")
	genesisFile = viper.GetString("backend.genesis")
	dbPath = viper.GetString("database.path")
	dbFile = viper.GetString("database.filename")
}

func createDatabase() (err error) {
	err = utils.CreateDir(datadir)
	if err != nil {
		return err
	}

	db, err = database.CreateDatabase(dbPath, dbFile)
	if err != nil {
		return err
	}
	defer db.Close()

	return nil
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

func defaultIPC() string {
	return filepath.Join(datadir, "/geth.ipc")
}
