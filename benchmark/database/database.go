package database

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/relab/bbchain-dapp/src/utils"
	bolt "go.etcd.io/bbolt"
)

type Database struct {
	*bolt.DB
	options  *bolt.Options
	filename string
}

var defaultBoltOptions = bolt.Options{Timeout: 1 * time.Second}

// CreateDatabase creates a new database
func CreateDatabase(dbpath, dbfile string) (*Database, error) {
	err := utils.CreateDir(dbpath)
	if err != nil {
		return nil, err
	}

	dbFileName := path.Join(dbpath, dbfile)
	return NewDatabase(dbFileName, &defaultBoltOptions)
}

// NewDatabase returns a new database instance
func NewDatabase(dbFilename string, opts *bolt.Options) (*Database, error) {
	db, err := bolt.Open(dbFilename, 0644, opts)
	if err != nil {
		return nil, err
	}
	return &Database{db, opts, dbFilename}, nil
}

// OpenDB open the boltdb instance
func (d *Database) OpenDB() (err error) {
	if d.DB, err = bolt.Open(d.filename, 0644, d.options); err != nil {
		return err
	}
	return err
}

// DeleteBucket deletes the last bucket at path
func (d *Database) DeleteBucket(path []string) error {
	s := len(path)
	if s < 1 {
		return fmt.Errorf("empty list of buckets")
	}
	err := d.Update(func(tx *bolt.Tx) error {
		if s == 1 { // root bucket
			tx.DeleteBucket([]byte(path[0]))
		}
		key := []byte(path[s-1])
		b, err := GetBucket(tx, path[:s-1])
		if b == nil {
			return err
		}
		if isBucket := b.Bucket(key); isBucket != nil {
			return b.DeleteBucket(key)
		}
		return err
	})
	return err
}

// CreateBucketPath create a list of buckets
func (d *Database) CreateBucketPath(path []string) error {
	if len(path) < 1 {
		return fmt.Errorf("empty list of buckets")
	}

	err := d.Update(func(tx *bolt.Tx) error {
		var err error
		b := tx.Bucket([]byte(path[0]))
		if b == nil {
			b, err = tx.CreateBucket([]byte(path[0]))
			if err != nil {
				return err
			}
		}

		for _, p := range path[1:] {
			next := b.Bucket([]byte(p))
			if next == nil {
				next, err = b.CreateBucket([]byte(p))
				if err != nil {
					return err
				}
			}
			b = next
		}
		return err
	})
	return err
}

// GetBucket returns the last bucket from the given path
func GetBucket(tx *bolt.Tx, path []string) (*bolt.Bucket, error) {
	b := tx.Bucket([]byte(path[0]))
	if b == nil {
		return nil, fmt.Errorf("bucket not found at path: " + path[0])
	}
	for i := 1; i < len(path); i++ {
		b = b.Bucket([]byte(path[i]))
		if b == nil {
			return nil, fmt.Errorf("bucket not found at path: " + strings.Join(path[:i+1], "/"))
		}
	}
	return b, nil
}

// GetKeys returns the list of keys at a given path, encoded as string
func (d *Database) GetKeys(path []string) (keys []string, err error) {
	if len(path) < 1 {
		return keys, fmt.Errorf("empty list of buckets")
	}

	err = d.View(func(tx *bolt.Tx) error {
		b, err := GetBucket(tx, path)
		if b == nil {
			return err
		}

		err = b.ForEach(func(k, v []byte) error {
			if v != nil { // buckets have nil value
				keys = append(keys, string(k))
			}
			return nil
		})
		return err
	})
	return keys, err
}

func (d *Database) AddEntry(path []string, key []byte, val []byte) error {
	err := d.Update(func(tx *bolt.Tx) error {
		b, err := GetBucket(tx, path)
		if err != nil {
			return err
		}
		err = b.Put(key, val)
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

// DeleteEntry deletes the entry on the path
func (d *Database) DeleteEntry(path []string, key []byte) error {
	err := d.Update(func(tx *bolt.Tx) error {
		b, err := GetBucket(tx, path)
		if err != nil {
			return err
		}
		return b.Delete(key)
	})
	return err
}

func (d *Database) GetEntry(path []string, key []byte) ([]byte, error) {
	var entry []byte
	err := d.View(func(tx *bolt.Tx) error {
		b, err := GetBucket(tx, path)
		if err != nil {
			return err
		}
		entry = b.Get(key)
		return nil
	})
	return entry, err
}

func (d *Database) MapValues(path []string, process func(value []byte) error) error {
	err := d.View(func(tx *bolt.Tx) error {
		b, err := GetBucket(tx, path)
		if err != nil {
			return err
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if v != nil {
				process(v)
			}
		}
		return nil
	})
	return err
}
