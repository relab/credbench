package database

import (
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

var (
	ErrBucketNotFound = fmt.Errorf("bucket not found")
	ErrNoBuckets      = fmt.Errorf("there is no buckets")
	ErrEntryNotFound  = fmt.Errorf("entry not found")
)

type boltDB struct {
	db      *bolt.DB
	options *bolt.Options
	path    string
}

var DefaultBoltOptions = bolt.Options{Timeout: 1 * time.Second}

// NewDatabase returns a new database instance
func NewDatabase(path string, opts *bolt.Options) (Database, error) {
	db, err := bolt.Open(path, 0644, opts)
	if err != nil {
		return nil, err
	}
	return &boltDB{db: db, options: opts, path: path}, nil
}

// OpenDB open the boltDB instance
func (d *boltDB) OpenDB() (err error) {
	if d.db, err = bolt.Open(d.path, 0644, d.options); err != nil {
		return err
	}
	return err
}

func (d *boltDB) Close() error {
	return d.db.Close()
}

// getBucket returns the last bucket from the given path
func getBucket(tx *bolt.Tx, path []string) (*bolt.Bucket, error) {
	b := tx.Bucket([]byte(path[0]))
	if b == nil {
		return nil, ErrBucketNotFound
	}
	for i := 1; i < len(path); i++ {
		b = b.Bucket([]byte(path[i]))
		if b == nil {
			return nil, ErrBucketNotFound
		}
	}
	return b, nil
}

// DeleteBucket deletes the last bucket at path
func (d *boltDB) DeleteBucket(path []string) error {
	s := len(path)
	if s < 1 {
		return ErrNoBuckets
	}
	err := d.db.Update(func(tx *bolt.Tx) error {
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

// CreateBucketPath create a list of buckets
func (d *boltDB) CreateBucketPath(path []string) error {
	if len(path) < 1 {
		return ErrNoBuckets
	}

	err := d.db.Update(func(tx *bolt.Tx) error {
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
func (d *boltDB) GetKeys(path []string) (keys [][]byte, err error) {
	if len(path) < 1 {
		return keys, ErrNoBuckets
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

func (d *boltDB) AddEntry(path []string, key []byte, value []byte) error {
	err := d.db.Update(func(tx *bolt.Tx) error {
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
func (d *boltDB) DeleteEntry(path []string, key []byte) error {
	err := d.db.Update(func(tx *bolt.Tx) error {
		b, err := getBucket(tx, path)
		if err != nil {
			return err
		}
		return b.Delete(key)
	})
	return err
}

func (d *boltDB) GetEntry(path []string, key []byte) ([]byte, error) {
	var entry []byte
	err := d.db.View(func(tx *bolt.Tx) error {
		b, err := getBucket(tx, path)
		if err != nil {
			return err
		}
		entry = b.Get(key)
		return nil
	})
	return entry, err
}

func (d *boltDB) MapValues(path []string, process func(value []byte) error) error {
	err := d.db.View(func(tx *bolt.Tx) error {
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

func (d *boltDB) GetFirstEntry(path []string) ([]byte, []byte, error) {
	var key, value []byte
	err := d.db.View(func(tx *bolt.Tx) error {
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
// allow the datastore to access the boltdb methods.
func (d *boltDB) GetNextEntry(path []string, key []byte) ([]byte, []byte, error) {
	var next, value []byte
	err := d.db.View(func(tx *bolt.Tx) error {
		b, err := getBucket(tx, path)
		if err != nil {
			return err
		}
		c := b.Cursor()
		k, _ := c.Seek(key)
		if k != nil {
			next, value = c.Next()
		} else {
			return ErrEntryNotFound
		}
		return nil
	})
	return next, value, err
}
