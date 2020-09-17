package database

type Database interface {
	OpenDB() error
	Close() error

	CreateBucketPath(path []string) error
	DeleteBucket(path []string) error

	AddEntry(path []string, key []byte, val []byte) error
	DeleteEntry(path []string, key []byte) error
	GetEntry(path []string, key []byte) ([]byte, error)

	GetKeys(path []string) (keys [][]byte, err error)
	GetFirstEntry(path []string) ([]byte, []byte, error)
	GetNextEntry(path []string, key []byte) ([]byte, []byte, error)
	MapValues(path []string, process func(value []byte) error) error
}
