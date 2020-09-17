package datastore

import "github.com/relab/ct-eth-dapp/benchmark/database"

type DataStore struct {
	db    database.Database
	sPath []string
}
