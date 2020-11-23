package datastore

import (
	"github.com/relab/ct-eth-dapp/bench/database"
)

type DataStore struct {
	db   *database.BoltDB
	path string
}
