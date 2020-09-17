package datastore

import (
	"fmt"

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
	return db.CreateBucketPath([]string{courseBucket})
}

func NewCourseStore(db database.Database, courseAddress common.Address) *CourseStore {
	return &CourseStore{
		course:  &DataStore{db: db, sPath: []string{courseBucket}},
		address: courseAddress,
	}
}

func (cs CourseStore) AddCourse(course *pb.Course) error {
	if course == nil {
		return fmt.Errorf("course cannot be nil")
	}
	buf, err := proto.Marshal(course)
	if err != nil {
		return err
	}
	return cs.course.db.AddEntry(cs.course.sPath, course.Address, buf)
}

func (cs CourseStore) GetCourse() (*pb.Course, error) {
	course := &pb.Course{}
	buf, err := cs.course.db.GetEntry(cs.course.sPath, cs.address.Bytes())
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
