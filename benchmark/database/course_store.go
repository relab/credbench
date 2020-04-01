package database

// Bucket("courses")
// - Bucket("course_address")
// -- Bucket("course_address/evaluators")
// kv: evaluator_address -> Account
// -- Bucket("course_address/students")
// kv: student_address -> Account
// -- Bucket("course_address/credentials")
// kv: student_address -> []digest
var (
	courseRootBucket  = "courses"
	evaluatorsBucket  = "evaluators"
	studentsBucket    = "students"
	credentialsBucket = "credentials"
)

var courseBuckets = []string{
	evaluatorsBucket,
	studentsBucket,
	credentialsBucket,
}

type CourseStore struct {
	DB   *Database
	Path []string
}

func CreateCourseStore(d *Database, courseAddress string) (*CourseStore, error) {
	var err error
	for _, b := range courseBuckets {
		err = d.CreateBucketPath([]string{courseRootBucket, courseAddress, b})
		if err != nil {
			return &CourseStore{}, err
		}
	}
	return NewCourseStore(d, []string{courseRootBucket, courseAddress}), err
}

func NewCourseStore(d *Database, path []string) *CourseStore {
	return &CourseStore{DB: d, Path: path}
}
