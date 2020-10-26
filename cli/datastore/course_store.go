package datastore

import (
	"github.com/ethereum/go-ethereum/common"
	proto "google.golang.org/protobuf/proto"

	"github.com/relab/ct-eth-dapp/cli/database"
	pb "github.com/relab/ct-eth-dapp/cli/proto"
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

func CreateCourseStore(db *database.BoltDB) error {
	return db.CreateBucketPath(courseBucket)
}

func NewCourseStore(db *database.BoltDB, courseAddress common.Address) *CourseStore {
	return &CourseStore{
		store:   &DataStore{db: db, path: courseBucket},
		address: courseAddress,
	}
}

func (cs *CourseStore) PutCourse(course *pb.Course) error {
	if course == nil {
		return ErrEmptyData
	}
	value, err := proto.Marshal(course)
	if err != nil {
		return err
	}
	address := common.HexToAddress(course.ContractAddress)
	if address == (common.Address{}) {
		return ErrZeroAddress
	}
	return cs.store.db.AddEntry(cs.store.path, address.Bytes(), value)
}

func (cs CourseStore) GetCourse() (*pb.Course, error) {
	course := &pb.Course{}
	buf, err := cs.store.db.GetEntry(cs.store.path, cs.address.Bytes())
	if err != nil {
		return nil, err
	}
	if buf != nil {
		err := proto.Unmarshal(buf, course)
		if err != nil {
			return nil, err
		}
	}
	return course, err
}

func (cs *CourseStore) SetStudents(students Accounts) error {
	course, err := cs.GetCourse()
	if err != nil {
		return err
	}
	course.Students = students
	return cs.PutCourse(course)
}

func (cs *CourseStore) AddCredential(credential *pb.Credential) error {
	course, err := cs.GetCourse()
	if err != nil {
		return err
	}
	course.Credentials = append(course.Credentials, credential)
	return cs.PutCourse(course)
}
