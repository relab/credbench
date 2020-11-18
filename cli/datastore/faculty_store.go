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

func (fs *FacultyStore) PutFaculty(faculty *pb.Faculty) error {
	if faculty == nil {
		return ErrEmptyData
	}
	address := common.BytesToAddress(faculty.Address)
	if address == (common.Address{}) {
		return ErrZeroAddress
	}
	value, err := proto.Marshal(faculty)
	if err != nil {
		return err
	}
	return fs.store.db.Put(fs.store.path, address.Bytes(), value)
}

func (fs *FacultyStore) GetFaculty() (*pb.Faculty, error) {
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
	if len(faculty.Students) == 0 {
		faculty.Students = make(map[string]*pb.Student)
	}
	return faculty, err
}

func (fs *FacultyStore) AddSemester(semester [32]byte) error {
	faculty, err := fs.GetFaculty()
	if err != nil {
		return err
	}

	faculty.Semesters = append(faculty.Semesters, semester[:])
	return fs.PutFaculty(faculty)
}

func (fs *FacultyStore) AddStudent(students ...common.Address) error {
	faculty, err := fs.GetFaculty()
	if err != nil {
		return err
	}

	for _, sAddr := range students {
		if _, ok := faculty.Students[sAddr.Hex()]; !ok {
			faculty.Students[sAddr.Hex()] = &pb.Student{StudentId: sAddr.Bytes()}
		}
	}
	return fs.PutFaculty(faculty)
}
