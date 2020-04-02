package cmd

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	bolt "go.etcd.io/bbolt"

	"github.com/relab/bbchain-dapp/benchmark/database"
	"github.com/relab/bbchain-dapp/src/core/client"
	"github.com/relab/bbchain-dapp/src/utils"
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

var (
	clientConn client.BBChainEthClient
	db         *database.Database
)

var rootCmd = &cobra.Command{
	Use:   "benchmark",
	Short: "BBChain benchmark generator",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		err := setupDB(dbPath, dbFile)
		if err != nil {
			log.Fatalln(err.Error())
		}

		clientConn, err = setupClient()
		if err != nil {
			log.Fatalln(err.Error())
		}
	},
	PersistentPostRun: func(_ *cobra.Command, _ []string) {
		clientConn.Close()
		db.Close()
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
	rootCmd.PersistentFlags().StringVar(&ipcFile, "ipc",
		defaultIPC(), "Ethereum Inter-process Communication file")
	rootCmd.PersistentFlags().BoolVar(&waitPeers, "waitPeers", false, "Minimum number of peers connected")
	rootCmd.PersistentFlags().StringVar(&dbPath, "dbPath", defaultDatabasePath(), "Path to the database file")
	rootCmd.PersistentFlags().StringVar(&dbFile, "dbFile", "bbchain.db", "File name of the database")
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

func setupClient() (client.BBChainEthClient, error) {
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

func setupDB(dbpath, dbfile string) (err error) {
	dbFileName := path.Join(dbpath, dbfile)
	db, err = database.NewDatabase(dbFileName, &bolt.Options{Timeout: 1 * time.Second})
	return err
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

func defaultDatabasePath() string {
	return filepath.Join(datadir, "database")
}

func defaultIPC() string {
	return filepath.Join(datadir, "/geth.ipc")
}
