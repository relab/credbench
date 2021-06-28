package datastore

import (
	"github.com/relab/credbench/bench/database"
)

type DataStore struct {
	db   *database.BoltDB
	path string
}
