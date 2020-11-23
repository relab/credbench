package cmd

import (
	"os"
	"os/user"
	"path"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	bolt "go.etcd.io/bbolt"

	"github.com/relab/ct-eth-dapp/cli/database"
	"github.com/relab/ct-eth-dapp/cli/datastore"
	"github.com/relab/ct-eth-dapp/cli/transactor"
	"github.com/relab/ct-eth-dapp/src/client"
	"github.com/relab/ct-eth-dapp/src/fileutils"

	pb "github.com/relab/ct-eth-dapp/cli/proto"
)

var (
	configFile     string
	backendURL     string
	defaultAccount string
	datadir        string
	logdir         string
	ipcFile        string
	waitPeers      bool
	testFile       string
	dbPath         string
	dbFile         string
	consensus      string
)

var (
	backend       *ethclient.Client
	db            *database.BoltDB
	accountStore  *datastore.EthAccountStore
	defaultSender common.Address
	executor      *transactor.Transactor
)

var rootCmd = &cobra.Command{
	Use:   "cli",
	Short: "Ethereum Credential Transparency System",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		err := setupDB(dbPath, dbFile)
		if err != nil {
			log.Fatal(err)
		}

		clientConn, err := setupClient()
		if err != nil {
			log.Fatal(err)
		}
		backend, _ = clientConn.Backend()
		// default transaction runner
		executor = transactor.NewTransactor(backend)
	},
	PersistentPostRun: func(_ *cobra.Command, _ []string) {
		db.Close()
		backend.Close()
	},
}

func Execute() {
	rootCmd.AddCommand(
		listenCmd,
		genesisCmd,
		newTestCmd(),
		newAccountCmd(),
		newDeployCmd(),
		newCourseCmd(),
		newFacultyCmd(),
		newVerifyCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&datadir, "datadir", defaultDatadir(), "path to the data directory")
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file")
	rootCmd.PersistentFlags().StringVar(&defaultAccount, "default_account", "", "Ethereum default account address")
	rootCmd.PersistentFlags().StringVar(&backendURL, "backendURL", "http://127.0.0.1:8545", "Blockchain backend host:port")
	rootCmd.PersistentFlags().StringVar(&ipcFile, "ipc", defaultIPC(), "Ethereum Inter-process Communication file")
	rootCmd.PersistentFlags().BoolVar(&waitPeers, "wait_peers", false, "Minimum number of peers connected")
	rootCmd.PersistentFlags().StringVar(&dbPath, "dbPath", "./database", "Path to the database file")
	rootCmd.PersistentFlags().StringVar(&dbFile, "dbFile", "cteth.db", "File name of the database")
	rootCmd.PersistentFlags().StringVar(&consensus, "consensus", "ethash", "Consensus engine: poa/ethash")
	rootCmd.PersistentFlags().StringVar(&testFile, "testFile", "test-config.json", "test case config file")

	cobra.OnInitialize(initConfig)

	log.SetOutput(os.Stderr)
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:          true,
		TimestampFormat:        time.RFC3339,
		DisableLevelTruncation: true,
	})
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
		log.Infoln("Using config file:", viper.ConfigFileUsed())
		parseConfigFile()
	}
	err := initSetup()
	if err != nil {
		log.Fatal(err)
	}
}

func initSetup() (err error) {
	err = fileutils.CreateDir(datadir)
	if err != nil {
		return err
	}
	err = fileutils.CreateDir(logdir)
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

func createDatabase() (err error) {
	err = fileutils.CreateDir(dbPath)
	if err != nil {
		return err
	}

	return nil
}

func setupDB(dbpath, dbfile string) (err error) {
	err = createDatabase()
	if err != nil {
		return err
	}

	db, err = database.NewDatabase(path.Join(dbpath, dbfile), &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}

	err = datastore.CreateEthAccountStore(db)
	if err != nil {
		return err
	}
	accountStore = datastore.NewEthAccountStore(db)

	err = datastore.CreateCourseStore(db)
	if err != nil {
		return err
	}

	err = datastore.CreateFacultyStore(db)
	if err != nil {
		return err
	}
	// TODO: initialize credential store
	return nil
}

func parseConfigFile() {
	datadir = viper.GetString("datadir")
	logdir = filepath.Join(datadir, "/logs")
	defaultAccount = viper.GetString("default_account")
	backendURL = "http://" + viper.GetString("backend.host") + ":" + viper.GetString("backend.port")
	ipcFile = viper.GetString("backend.ipc")
	waitPeers = viper.GetBool("backend.wait_peers")
	consensus = viper.GetString("chain.consensus")
	dbPath = viper.GetString("database.path")
	dbFile = viper.GetString("database.filename")
}

func defaultConfigPath() string {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
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

func defaultIPC() string {
	return filepath.Join(datadir, "/geth.ipc")
}

func loadDefaultAccount() error {
	if defaultAccount == "" {
		accounts, err := accountStore.GetByType(1, pb.Type_DEPLOYER)
		if err != nil || len(accounts) == 0 {
			if err == datastore.ErrNoAccountsFound {
				accounts, err = accountStore.GetAndSelect(1, pb.Type_DEPLOYER)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
		defaultSender = common.BytesToAddress(accounts[0].GetAddress())
		defaultAccount = defaultSender.Hex()
	} else {
		defaultSender = common.HexToAddress(defaultAccount)
	}
	log.Info("Using sender account: ", defaultAccount)

	return nil
}
