package schemes

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"time"

	proto "github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/r0qs/bbchain-dapp/src/core/backends"
)

var (
	admsAccounts       backends.Accounts
	evaluatorsAccounts backends.Accounts
	studentsAccounts   backends.Accounts
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	// 10 accounts on total
	admsAccounts = backends.TestAccounts[:2]
	evaluatorsAccounts = backends.TestAccounts[2:4]
	studentsAccounts = backends.TestAccounts[4:]
}

func hashString(s string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
}

func NewFakeAssignmentGrade(teacherID, studentID string) *AssignmentGrade {
	c := rand.Intn(100)
	return &AssignmentGrade{
		Id:          hashString(fmt.Sprintf("%s%d", "AssignmentFile-", c)),
		Name:        fmt.Sprintf("%s%d", "Exam ", c),
		Code:        fmt.Sprintf("%s%d", "EX-", c),
		Category:    "InternalActivity",
		Type:        []string{"MandatoryActivity"},
		Language:    "en",
		Description: "This is an exam description",
		Evaluators: []*Entity{
			&Entity{
				Id: teacherID,
			},
		},
		Student: &Entity{
			Id: studentID,
		},
		Grade:           int64(c),
		StudentPresence: "Physical",
	}
}

func NewFakeAssignmentGradeCredential(teacherID string, courseEntity *Entity, ag *AssignmentGrade) *AssignmentGradeCredential {
	creationTime := time.Now().UTC().UnixNano()
	return &AssignmentGradeCredential{
		Assignment:       ag,
		CreatedBy:        teacherID,
		CreatedAt:        &timestamp.Timestamp{Seconds: creationTime},
		OfferedBy:        []*Entity{courseEntity},
		EvidenceDocument: "use swarm hash here",
		DocumentPresence: "Physical",
	}
}

// TODO: DRY
func (ag *AssignmentGradeCredential) Hash() [32]byte {
	data, _ := proto.Marshal(ag)
	return sha256.Sum256(data)
}

// n is the number of assignments to generate
func generateFakeAssignmentGrades(n int) (assignments []*AssignmentGrade) {
	t := len(evaluatorsAccounts)
	studentID := studentsAccounts[rand.Intn(len(studentsAccounts))].Address.Hex()
	for i := 0; i < n; i++ {
		teacherID := evaluatorsAccounts[rand.Intn(t)].Address.Hex()
		ag := NewFakeAssignmentGrade(teacherID, studentID)
		assignments = append(assignments, ag)
	}
	return assignments
}

func generateFakeAssignmentGradeCredentials(courseID string, assignments []*AssignmentGrade) (credentials []*AssignmentGradeCredential) {
	courseEntity := &Entity{
		Id:   courseID,
		Name: "Course Test Contract",
	}
	for _, a := range assignments {
		teacherID := a.Evaluators[0].GetId() // the first evaluator will be the creator of the assignment
		c := NewFakeAssignmentGradeCredential(teacherID, courseEntity, a)
		credentials = append(credentials, c)
	}
	return credentials
}

// TODO: DRY
func NewFakeCourseGrade(courseID string, credentials []*AssignmentGradeCredential) *CourseGrade {
	firstTimestamp := credentials[0].CreatedAt
	lastTimestamp := credentials[len(credentials)-1].CreatedAt
	var d duration.Duration
	d.Seconds = lastTimestamp.Seconds - firstTimestamp.Seconds
	d.Nanos = lastTimestamp.Nanos - firstTimestamp.Nanos

	teacherID := credentials[0].CreatedBy
	studentID := credentials[0].Assignment.GetStudent().GetId()
	c := rand.Intn(100)

	cgrade := &CourseGrade{
		Id:          hashString(fmt.Sprintf("%s%d", "CourseFinalGrade-", c)),
		Name:        fmt.Sprintf("%s%d", "Course ", c),
		Code:        fmt.Sprintf("%s%d", "CS-", c),
		Category:    "InternalCourse",
		Type:        []string{"MandatoryCourse"},
		Language:    "en",
		Description: "The course provides insight into both theoretical and practical aspects of something",
		Duration:    &d,
		Teachers: []*Entity{
			&Entity{
				Id: teacherID,
			},
		},
		Evaluators: []*Entity{
			&Entity{
				Id: teacherID,
			},
		},
		Student: &Entity{
			Id: studentID,
		},
		GradingSystem:   "ECTS",
		TotalCredits:    20,
		FinalGrade:      int64(c),
		StudentPresence: "Physical",
		Assignments:     credentials,
	}
	return cgrade
}

func NewFakeCourseGradeCredential(teacherID string, facultyEntity *Entity, cg *CourseGrade) *CourseGradeCredential {
	creationTime := time.Now().UTC().UnixNano()
	return &CourseGradeCredential{
		Course:           cg,
		CreatedBy:        teacherID,
		CreatedAt:        &timestamp.Timestamp{Seconds: creationTime},
		OfferedBy:        []*Entity{facultyEntity},
		EvidenceDocument: "use swarm hash here",
		DocumentPresence: "Physical",
	}
}

func generateFakeCoursesGrade(coursesIDS []string) (courses []*CourseGrade) {
	t := len(evaluatorsAccounts)
	for _, courseID := range coursesIDS {
		assignments := generateFakeAssignmentGrades(rand.Intn(3))
		credentials := generateFakeAssignmentGradeCredentials(courseID, assignments)
		teacherID := evaluatorsAccounts[rand.Intn(t)].Address.Hex()
		c := NewFakeCourseGrade(teacherID, credentials)
		courses = append(courses, c)
	}
	return courses
}

func generateFakeCoursesGradeCredentials(facultyID string, courses []*CourseGrade) (credentials []*CourseGradeCredential) {
	facultyEntity := &Entity{
		Id:   facultyID,
		Name: "Faculty Test Contract",
	}
	for _, c := range courses {
		teacherID := c.Evaluators[0].GetId()
		c := NewFakeCourseGradeCredential(teacherID, facultyEntity, c)
		credentials = append(credentials, c)
	}
	return credentials
}

func (cg *CourseGradeCredential) Hash() [32]byte {
	data, _ := proto.Marshal(cg)
	return sha256.Sum256(data)
}
