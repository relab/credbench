package datastore

import (
	"github.com/relab/ct-eth-dapp/benchmark/database"
)

// FIXME: use string path
type DataStore struct {
	db    database.Database
	sPath []string
}
