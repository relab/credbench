package datastore

import (
	"github.com/ethereum/go-ethereum/common"
	proto "google.golang.org/protobuf/proto"

	"github.com/relab/ct-eth-dapp/bench/database"
	pb "github.com/relab/ct-eth-dapp/bench/proto"
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
	address := common.BytesToAddress(course.Address)
	if address == (common.Address{}) {
		return ErrZeroAddress
	}
	value, err := proto.Marshal(course)
	if err != nil {
		return err
	}
	return cs.store.db.Put(cs.store.path, address.Bytes(), value)
}

func (cs CourseStore) GetCourse() (*pb.Course, error) {
	course := &pb.Course{}
	buf, err := cs.store.db.Get(cs.store.path, cs.address.Bytes())
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

// SetStudents override the current students of the course
func (cs *CourseStore) SetStudents(students Accounts) error {
	course, err := cs.GetCourse()
	if err != nil {
		return err
	}
	course.Students = students.ToBytes()
	return cs.PutCourse(course)
}
