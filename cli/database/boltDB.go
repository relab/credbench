package database

import (
	"errors"
	"strings"

	bolt "go.etcd.io/bbolt"
)

var (
	errBucketNotFound    = errors.New("bucket not found")
	errInvalidBucketPath = errors.New("invalid path")
	errEntryNotFound     = errors.New("entry not found")
)

type BoltDB struct {
	db      *bolt.DB
	options *bolt.Options
	path    string
}

func normalizePath(s string) ([]string, error) {
	path := strings.Split(strings.Trim(s, "/"), "/")
	if len(path) < 1 {
		return nil, errInvalidBucketPath
	}
	return path, nil
}

// NewDatabase returns a new database instance
func NewDatabase(path string, opts *bolt.Options) (*BoltDB, error) {
	db, err := bolt.Open(path, 0644, opts)
	if err != nil {
		return nil, err
	}
	return &BoltDB{db: db, options: opts, path: path}, nil
}

// OpenDB open the BoltDB instance
func (d *BoltDB) OpenDB() (err error) {
	if d.db, err = bolt.Open(d.path, 0644, d.options); err != nil {
		return err
	}
	return err
}

func (d *BoltDB) Close() error {
	return d.db.Close()
}

// getBucket returns the last bucket from the given path
//FIXME use string path
func getBucket(tx *bolt.Tx, path []string) (*bolt.Bucket, error) {
	if len(path) < 1 {
		return nil, errInvalidBucketPath
	}
	b := tx.Bucket([]byte(path[0]))
	if b == nil {
		return nil, errBucketNotFound
	}
	for i := 1; i < len(path); i++ {
		b = b.Bucket([]byte(path[i]))
		if b == nil {
			return nil, errBucketNotFound
		}
	}
	return b, nil
}

// DeleteBucket deletes the last bucket at path
func (d *BoltDB) DeleteBucket(pathStr string) error {
	path, err := normalizePath(pathStr)
	if err != nil {
		return err
	}
	s := len(path)
	err = d.db.Update(func(tx *bolt.Tx) error {
		if s == 1 { // root bucket
			tx.DeleteBucket([]byte(path[0]))
		}
		key := []byte(path[s-1])
		b, err := getBucket(tx, path[:s-1])
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

// BucketExists check if a bucket exists
func (d *BoltDB) BucketExists(pathStr string) bool {
	path, err := normalizePath(pathStr)
	if err != nil || len(path) < 1 {
		return false
	}
	var b *bolt.Bucket
	err = d.db.View(func(tx *bolt.Tx) error {
		b, err = getBucket(tx, path)
		if err != nil {
			return err
		}
		return nil
	})
	return b != nil && err == nil
}

// CreateBucketPath create a list of buckets
func (d *BoltDB) CreateBucketPath(pathStr string) error {
	path, err := normalizePath(pathStr)
	if err != nil {
		return err
	}
	err = d.db.Update(func(tx *bolt.Tx) error {
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

// GetKeys returns the list of keys at a given path, encoded as string
func (d *BoltDB) GetKeys(pathStr string) (keys [][]byte, err error) {
	path, err := normalizePath(pathStr)
	if err != nil {
		return keys, err
	}
	err = d.db.View(func(tx *bolt.Tx) error {
		b, err := getBucket(tx, path)
		if b == nil {
			return err
		}

		err = b.ForEach(func(k, v []byte) error {
			if v != nil { // buckets have nil value
				keys = append(keys, k)
			}
			return nil
		})
		return err
	})
	return keys, err
}

func (d *BoltDB) AddEntry(pathStr string, key []byte, value []byte) error {
	path, err := normalizePath(pathStr)
	if err != nil {
		return err
	}
	err = d.db.Update(func(tx *bolt.Tx) error {
		b, err := getBucket(tx, path)
		if err != nil {
			return err
		}
		err = b.Put(key, value)
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

// DeleteEntry deletes the entry on the path
func (d *BoltDB) DeleteEntry(pathStr string, key []byte) error {
	path, err := normalizePath(pathStr)
	if err != nil {
		return err
	}
	err = d.db.Update(func(tx *bolt.Tx) error {
		b, err := getBucket(tx, path)
		if err != nil {
			return err
		}
		return b.Delete(key)
	})
	return err
}

func (d *BoltDB) GetEntry(pathStr string, key []byte) (entry []byte, err error) {
	path, err := normalizePath(pathStr)
	if err != nil {
		return entry, err
	}
	err = d.db.View(func(tx *bolt.Tx) error {
		b, err := getBucket(tx, path)
		if err != nil {
			return err
		}
		entry = b.Get(key)
		return nil
	})
	return entry, err
}

func (d *BoltDB) MapValues(pathStr string, process func(value []byte) error) error {
	path, err := normalizePath(pathStr)
	if err != nil {
		return err
	}
	err = d.db.View(func(tx *bolt.Tx) error {
		b, err := getBucket(tx, path)
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

func (d *BoltDB) GetFirstEntry(pathStr string) (key []byte, value []byte, err error) {
	path, err := normalizePath(pathStr)
	if err != nil {
		return key, value, err
	}
	err = d.db.View(func(tx *bolt.Tx) error {
		b, err := getBucket(tx, path)
		if err != nil {
			return err
		}
		c := b.Cursor()
		if c != nil {
			key, value = c.First()
		}
		return nil
	})
	return key, value, err
}

// FIXME: not efficient, open a db tx every time...consider to
// allow the datastore to access the BoltDB methods.
func (d *BoltDB) GetNextEntry(pathStr string, key []byte) (next []byte, value []byte, err error) {
	path, err := normalizePath(pathStr)
	if err != nil {
		return next, value, err
	}
	err = d.db.View(func(tx *bolt.Tx) error {
		b, err := getBucket(tx, path)
		if err != nil {
			return err
		}
		c := b.Cursor()
		k, _ := c.Seek(key)
		if k != nil {
			next, value = c.Next()
		} else {
			return errEntryNotFound
		}
		return nil
	})
	return next, value, err
}
