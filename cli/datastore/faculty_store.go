package datastore

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/relab/ct-eth-dapp/cli/database"
	pb "github.com/relab/ct-eth-dapp/cli/proto"
	proto "google.golang.org/protobuf/proto"
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

func (cs *FacultyStore) AddFaculty(faculty *pb.Faculty) error {
	if faculty == nil {
		return ErrEmptyData
	}
	value, err := proto.Marshal(faculty)
	if err != nil {
		return err
	}
	address := common.HexToAddress(faculty.ContractAddress)
	if address == (common.Address{}) {
		return ErrZeroAddress
	}
	return cs.store.db.Put(cs.store.path, address.Bytes(), value)
}

func (fs FacultyStore) GetFaculty() (*pb.Faculty, error) {
	faculty := &pb.Faculty{}
	buf, err := fs.store.db.Get(fs.store.path, fs.address.Bytes())
	if err != nil {
		return nil, err
	}
	if buf != nil {
		err := proto.Unmarshal(buf, faculty)
		if err != nil {
			return nil, err
		}
	}
	return faculty, err
}

func (fs *FacultyStore) SetCourses(courses []*pb.Course) error {
	if len(courses) == 0 {
		return nil
	}

	faculty, err := fs.GetFaculty()
	if err != nil {
		return err
	}

	faculty.Courses = courses
	return fs.AddFaculty(faculty)
}
