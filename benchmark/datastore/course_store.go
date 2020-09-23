package datastore

import (
	"github.com/ethereum/go-ethereum/common"
	proto "github.com/golang/protobuf/proto"
	"github.com/relab/ct-eth-dapp/benchmark/database"
	pb "github.com/relab/ct-eth-dapp/benchmark/proto"
)

// Bucket("courses")
// kv: course_address -> CourseProto
var (
	courseBucket = "courses"
)

type CourseStore struct {
	store   *DataStore
	address common.Address
}

func CreateCourseStore(db database.Database) error {
	return db.CreateBucketPath(courseBucket)
}

func NewCourseStore(db database.Database, courseAddress common.Address) *CourseStore {
	return &CourseStore{
		store:   &DataStore{db: db, path: courseBucket},
		address: courseAddress,
	}
}

func (cs CourseStore) PutCourse(store *pb.Course) error {
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

func (cs CourseStore) GetCourse() (*pb.Course, error) {
	store := &pb.Course{}
	buf, err := cs.store.db.GetEntry(cs.store.path, cs.address.Bytes())
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

func (cs CourseStore) SetStudents(students Accounts) error {
	store, err := cs.GetCourse()
	if err != nil {
		return err
	}
	store.Students = students
	return cs.PutCourse(store)
}

func (cs CourseStore) AddCredential(credential *pb.Credential) error {
	store, err := cs.GetCourse()
	if err != nil {
		return err
	}
	store.Credentials = append(store.Credentials, credential)
	return cs.PutCourse(store)
}
