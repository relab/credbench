package datastore

import (
	"github.com/ethereum/go-ethereum/common"
	proto "github.com/golang/protobuf/proto"
	"github.com/relab/ct-eth-dapp/cli/database"
	pb "github.com/relab/ct-eth-dapp/cli/proto"
)

// Bucket("faculties")
// - Bucket("faculty_address")
// kv: faculty_address -> FacultyProto
var (
	facultyBucket = "faculties"
)

type FacultyStore struct {
	store   *DataStore
	address common.Address
}

func CreateFacultyStore(db *database.BoltDB) error {
	return db.CreateBucketPath(facultyBucket)
}

func NewFacultyStore(db *database.BoltDB, facultyAddress common.Address) *FacultyStore {
	return &FacultyStore{
		store:   &DataStore{db: db, path: facultyBucket},
		address: facultyAddress,
	}
}

func (cs FacultyStore) AddFaculty(store *pb.Faculty) error {
	if store == nil {
		return ErrEmptyData
	}
	value, err := proto.Marshal(store)
	if err != nil {
		return err
	}
	address := common.HexToAddress(store.ContractAddress)
	if address == (common.Address{}) {
		return errZeroAddress
	}
	return cs.store.db.AddEntry(cs.store.path, address.Bytes(), value)
}

func (fs FacultyStore) GetFaculty() (*pb.Faculty, error) {
	store := &pb.Faculty{}
	buf, err := fs.store.db.GetEntry(fs.store.path, fs.address.Bytes())
	if err != nil {
		return nil, err
	}
	if buf != nil {
		err := proto.Unmarshal(buf, store)
		if err != nil {
			return nil, err
		}
	}
	return store, err
}

func (fs FacultyStore) SetCourses(courses []*pb.Course) error {
	store, err := fs.GetFaculty()
	if err != nil {
		return err
	}
	store.Courses = courses
	return fs.AddFaculty(store)
}
