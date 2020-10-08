package datastore

import (
	"github.com/relab/ct-eth-dapp/cli/database"
)

type DataStore struct {
	db   *database.BoltDB
	path string
}
