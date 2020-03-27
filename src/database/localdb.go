package database

import (
	"path"
	"time"

	bolt "go.etcd.io/bbolt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/relab/bbchain-dapp/src/utils"
)

// Database is a localdatabase for benchmarks
// It stores private keys and address information of deployed
// contracts and credentials created
type Database struct {
	*bolt.DB
}

func CreateDatabase(dbpath, dbfile string) (*Database, error) {
	err := utils.CreateDir(dbpath)
	if err != nil {
		return nil, err
	}

	dbFileName := path.Join(dbpath, dbfile)
	return NewDatabase(dbFileName)
}

func NewDatabase(dbFileName string) (*Database, error) {
	db, err := bolt.Open(dbFileName, 0644, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	return &Database{db}, nil
}

func (db *Database) CreateDBEntry(bucket string, entry *Entry) (err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(bucket))
		if err != nil {
			return err
		}

		err = b.Put(hexutil.MustDecode(entry.ID), entry.Serialize())
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func (db *Database) DeleteEntry(bucket string, id string) (err error) {
	key, err := hexutil.Decode(id)
	if err != nil {
		return err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		err = b.Delete(key)
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func (db *Database) DeleteEntryElement(bucket string, id string, elem string) (err error) {
	e, err := db.ReadUnMarshallEntry(bucket, id)
	if err != nil {
		return err
	}
	e.deleteElement(elem)
	return db.UpdateDBEntry(bucket, e)
}

func (db *Database) UpdateEntry(bucket string, id string, elements map[string][]string) error {
	e, err := db.ReadUnMarshallEntry(bucket, id)
	if err != nil {
		return err
	}
	e.addElement(elements)
	return db.UpdateDBEntry(bucket, e)
}

func (db *Database) UpdateDBEntry(bucket string, e *Entry) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		err := b.Put(hexutil.MustDecode(e.ID), e.Serialize())
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func (db *Database) ReadUnMarshallEntry(bucket string, id string) (*Entry, error) {
	key, err := hexutil.Decode(id)
	if err != nil {
		return nil, err
	}

	entry, err := db.ReadEntry(bucket, key)
	if err != nil {
		return nil, err
	}
	return DeserializeEntry(entry), nil
}

func (db *Database) ReadEntry(bucket string, key []byte) (entry []byte, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		entry = b.Get(key)
		return nil
	})
	return entry, err
}
