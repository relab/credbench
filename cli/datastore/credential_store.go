package datastore

import "github.com/relab/ct-eth-dapp/cli/database"

// Bucket("credentials")
// kv: digest -> Credential
var (
	credentialsBucket = "credentials"
)

type CredentialStore struct {
	store  *DataStore
	digest []byte
}

func CreateCredentialStore(db *database.BoltDB) error {
	if !db.BucketExists(credentialsBucket) {
		return db.CreateBucketPath(credentialsBucket)
	}
	return nil
}

func NewCredentialStore(db *database.BoltDB, digest []byte) *CredentialStore {
	return &CredentialStore{
		store:  &DataStore{db: db, path: credentialsBucket},
		digest: digest,
	}
}
