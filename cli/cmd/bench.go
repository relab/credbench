package cmd

import (
	"fmt"
	"log"
	"path"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	bolt "go.etcd.io/bbolt"

	"github.com/relab/ct-eth-dapp/cli/database"
	"github.com/relab/ct-eth-dapp/cli/datastore"
	"github.com/relab/ct-eth-dapp/src/fileutils"
)

var (
	testFile  string
	dbPath    string
	dbFile    string
	consensus string
)

var (
	db           *database.BoltDB
	accountStore *datastore.EthAccountStore
)

var benchCmd = &cobra.Command{
	Use:   "bench",
	Short: "Benchmark generator",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		parseBenchConfig()
		createDatabase()
		err := setupDB(dbPath, dbFile)
		if err != nil {
			log.Fatalln(err.Error())
		}

		fmt.Println("DATA SETUP")
		fmt.Println(dbPath, dbFile)
		clientConn, err := setupClient()
		if err != nil {
			log.Fatalln(err.Error())
		}
		backend, _ = clientConn.Backend()
	},
	PersistentPostRun: func(_ *cobra.Command, _ []string) {
		db.Close()
		backend.Close()
	},
}

func newBenchCmd() *cobra.Command {
	benchCmd.AddCommand(
		genesisCmd,
		generateCmd,
		newAccountsCmd(),
	)
	return benchCmd
}

func init() {
	benchCmd.PersistentFlags().StringVar(&dbPath, "dbPath", "./database", "Path to the database file")
	benchCmd.PersistentFlags().StringVar(&dbFile, "dbFile", "bench.db", "File name of the database")
	benchCmd.PersistentFlags().StringVar(&consensus, "consensus", "ethash", "Consensus engine: poa/ethash")
	benchCmd.PersistentFlags().StringVar(&testFile, "testFile", "test-config.json", "test case config file")
}

func parseBenchConfig() {
	path := defaultConfigPath()
	viper.AddConfigPath(path)
	viper.SetConfigName("bench-config")
	viper.ReadInConfig()
	fmt.Println("Using bench config file:", viper.ConfigFileUsed())
	consensus = viper.GetString("chain.consensus")
	dbPath = viper.GetString("database.path")
	dbFile = viper.GetString("database.filename")
	fmt.Printf("Using %s Consensus\n", consensus)
}

func createDatabase() (err error) {
	err = fileutils.CreateDir(dbPath)
	if err != nil {
		return err
	}

	return nil
}

func setupDB(dbpath, dbfile string) (err error) {
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
	//TODO: initialize credential store
	return nil
}
