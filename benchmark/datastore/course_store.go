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
	course  *DataStore
	address common.Address
}

func CreateCourseStore(db database.Database) error {
	return db.CreateBucketPath(courseBucket)
}

func NewCourseStore(db database.Database, courseAddress common.Address) *CourseStore {
	return &CourseStore{
		course:  &DataStore{db: db, path: courseBucket},
		address: courseAddress,
	}
}

func (cs CourseStore) PutCourse(course *pb.Course) error {
	if course == nil {
		return ErrEmptyData
	}
	value, err := proto.Marshal(course)
	if err != nil {
		return err
	}
	address := common.HexToAddress(course.ContractAddress)
	if address == (common.Address{}) {
		return errZeroAddress
	}
	return cs.course.db.AddEntry(cs.course.path, address.Bytes(), value)
}

func (cs CourseStore) GetCourse() (*pb.Course, error) {
	course := &pb.Course{}
	buf, err := cs.course.db.GetEntry(cs.course.path, cs.address.Bytes())
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

func (cs CourseStore) SetStudents(students Accounts) error {
	course, err := cs.GetCourse()
	if err != nil {
		return err
	}
	course.Students = students
	return cs.PutCourse(course)
}

func (cs CourseStore) AddCredential(credential *pb.Credential) error {
	course, err := cs.GetCourse()
	if err != nil {
		return err
	}
	course.Credentials = append(course.Credentials, credential)
	return cs.PutCourse(course)
}
