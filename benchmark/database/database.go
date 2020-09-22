package database

type Database interface {
	OpenDB() error
	Close() error

	CreateBucketPath(pathStr string) error
	DeleteBucket(pathStr string) error

	AddEntry(pathStr string, key []byte, val []byte) error
	DeleteEntry(pathStr string, key []byte) error
	GetEntry(pathStr string, key []byte) ([]byte, error)

	GetKeys(pathStr string) (keys [][]byte, err error)
	GetFirstEntry(pathStr string) ([]byte, []byte, error)
	GetNextEntry(pathStr string, key []byte) ([]byte, []byte, error)
	MapValues(pathStr string, process func(value []byte) error) error
}
