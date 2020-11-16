package database

import (
	"bytes"
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
// FIXME use string path
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
			err = tx.DeleteBucket([]byte(path[0]))
			if err != nil {
				return err
			}
		}
		key := []byte(path[s-1])
		var b *bolt.Bucket
		b, err = getBucket(tx, path[:s-1])
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

func (d *BoltDB) Keys(pathStr string) (keys [][]byte, err error) {
	return d.GetKeysWith(pathStr, func(v []byte) bool { return true })
}

// Keys returns the list of keys at a given path, encoded as string
func (d *BoltDB) GetKeysWith(pathStr string, filter func(value []byte) bool) (keys [][]byte, err error) {
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
			if v != nil && filter(v) { // buckets have nil value
				key := make([]byte, len(k))
				copy(key, k)
				keys = append(keys, key)
			}
			return nil
		})
		return err
	})
	return keys, err
}

func (d *BoltDB) Put(pathStr string, key []byte, value []byte) error {
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

// Delete deletes the entry on the path
func (d *BoltDB) Delete(pathStr string, key []byte) error {
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

func (d *BoltDB) Get(pathStr string, key []byte) (entry []byte, err error) {
	path, err := normalizePath(pathStr)
	if err != nil {
		return nil, err
	}
	if err = d.db.View(func(tx *bolt.Tx) error {
		b, err := getBucket(tx, path)
		if err != nil {
			return err
		}
		entry = b.Get(key)
		return nil
	}); err != nil {
		return nil, err
	}
	return entry, err
}

func (d *BoltDB) IterValues(pathStr string, process func(value []byte) error) error {
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
				err = process(v)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	return err
}

func (d *BoltDB) GetFirstEntry(pathStr string) (key []byte, value []byte, err error) {
	path, err := normalizePath(pathStr)
	if err != nil {
		return nil, nil, err
	}
	if err = d.db.View(func(tx *bolt.Tx) error {
		b, err := getBucket(tx, path)
		if err != nil {
			return err
		}
		c := b.Cursor()
		if c != nil {
			key, value = c.First()
		}
		return nil
	}); err != nil {
		return nil, nil, err
	}
	return key, value, err
}

// FIXME: not efficient, open a db tx every time...consider to
// allow the datastore to access the BoltDB methods.
func (d *BoltDB) GetNextEntry(pathStr string, key []byte) (next []byte, value []byte, err error) {
	path, err := normalizePath(pathStr)
	if err != nil {
		return nil, nil, err
	}
	if err = d.db.View(func(tx *bolt.Tx) error {
		b, err := getBucket(tx, path)
		if err != nil {
			return err
		}
		c := b.Cursor()
		k, _ := c.Seek(key)
		if k != nil {
			next, value = c.Next()
			return nil
		}
		return errEntryNotFound
	}); err != nil {
		return nil, nil, err
	}
	return next, value, err
}

func (d *BoltDB) IndexRead(pathStr string, prefix []byte, n int) ([][]byte, error) {
	var keys [][]byte

	path, err := normalizePath(pathStr)
	if err != nil {
		return keys, err
	}

	err = d.db.View(func(tx *bolt.Tx) error {
		b, err := getBucket(tx, path)
		if err != nil {
			return err
		}
		c := b.Cursor()
		i := 0
		for k, v := c.Seek(prefix); k != nil && bytes.Contains(k, prefix); k, v = c.Next() {
			if i == n {
				break
			}

			if v != nil {
				key := make([]byte, len(k))
				copy(key, k)
				keys = append(keys, key)
				i++
			}
		}

		return nil
	})
	return keys, err
}

func (d *BoltDB) Iterate(pathStr string, n int, conditionFn func(v []byte) (bool, error)) error {
	path, err := normalizePath(pathStr)
	if err != nil {
		return err
	}
	err = d.db.Batch(func(tx *bolt.Tx) error {
		b, err := getBucket(tx, path)
		if err != nil {
			return err
		}
		return iterate(b, n, conditionFn)
	})
	return err
}

func iterate(b *bolt.Bucket, n int, conditionFn func(v []byte) (bool, error)) error {
	var key, value []byte
	c := b.Cursor()
	if c != nil {
		key, value = c.First()
	}

	i := 0
	for i < n {
		ok, err := conditionFn(value)
		if ok {
			i++
		}
		if err != nil {
			return err
		}

		// Get next entry
		k, _ := c.Seek(key)
		if k != nil {
			key, value = c.Next()
		}
		if err != nil || key == nil {
			return err
		}
	}
	return nil
}

func (d *BoltDB) Update(pathStr string, key []byte, updateFn func(v []byte) ([]byte, error)) error {
	path, err := normalizePath(pathStr)
	if err != nil {
		return err
	}
	err = d.db.Batch(func(tx *bolt.Tx) error {
		b, err := getBucket(tx, path)
		if err != nil {
			return err
		}
		return update(b, key, b.Get(key), updateFn)
	})
	return err
}

func update(b *bolt.Bucket, key []byte, value []byte, updateFn func(v []byte) ([]byte, error)) error {
	updatedValue, err := updateFn(value)
	if err != nil {
		return err
	}

	err = b.Put(key, updatedValue)
	if err != nil {
		return err
	}
	return nil
}

func (d *BoltDB) Map(pathStr string, n int, applyIfFn func(v []byte) (bool, []byte, error)) error {
	path, err := normalizePath(pathStr)
	if err != nil {
		return err
	}
	err = d.db.Batch(func(tx *bolt.Tx) error {
		b, err := getBucket(tx, path)
		if err != nil {
			return err
		}
		return mapFn(b, n, applyIfFn)
	})
	return err
}

func mapFn(b *bolt.Bucket, n int, applyIfFn func(v []byte) (bool, []byte, error)) error {
	var key, value []byte
	c := b.Cursor()
	if c != nil {
		key, value = c.First()
	}

	i := 0
	for i < n {
		ok, updatedValue, err := applyIfFn(value)
		if ok && len(updatedValue) != 0 {
			err = b.Put(key, updatedValue)
			if err != nil {
				return err
			}
			i++
		}
		if err != nil {
			return err
		}

		// Get next entry
		k, _ := c.Seek(key)
		if k != nil {
			key, value = c.Next()
		}
		if err != nil || key == nil {
			return err
		}
	}
	return nil
}
