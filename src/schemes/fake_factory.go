package schemes

import (
	"fmt"
	"math/rand"
	"time"

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

func NewFakeAssignmentGrade(teacherID, studentID string) *AssignmentGrade {
	c := rand.Intn(100)
	return &AssignmentGrade{
		Id:          HashString(fmt.Sprintf("%s%d", "AssignmentFile-", c)),
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

func NewFakeAssignmentGradeCredential(creatorID string, courseEntity *Entity, ag *AssignmentGrade) *AssignmentGradeCredential {
	creationTime := time.Now().UTC().UnixNano()
	return &AssignmentGradeCredential{
		Assignment:       ag,
		CreatedBy:        creatorID,
		CreatedAt:        &timestamp.Timestamp{Seconds: creationTime},
		OfferedBy:        []*Entity{courseEntity},
		EvidenceDocument: "use swarm hash here",
		DocumentPresence: "Physical",
	}
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
	cgrade := &CourseGrade{
		Id:              courseID,
		Name:            fmt.Sprintf("%s-%s", "Course", courseID),
		Code:            fmt.Sprintf("%s%s", "C", courseID[:4]),
		Category:        "InternalCourse",
		Type:            []string{"MandatoryCourse"},
		Language:        "en",
		Description:     "The course provides insight into both theoretical and practical aspects of something",
		Duration:        &d,
		Teachers:        credentials[0].Assignment.GetEvaluators(),
		Evaluators:      credentials[0].Assignment.GetEvaluators(),
		Student:         credentials[0].Assignment.GetStudent(),
		GradingSystem:   "ECTS",
		TotalCredits:    20,
		FinalGrade:      int64(rand.Intn(100)), // TODO: compute base on assignments' grades
		StudentPresence: "Physical",
		Assignments:     credentials, //TODO: store only map with ids and grades
	}
	return cgrade
}

func NewFakeCourseGradeCredential(creatorID string, cg *CourseGrade) *CourseGradeCredential {
	creationTime := time.Now().UTC().UnixNano()
	return &CourseGradeCredential{
		Course:    cg,
		CreatedBy: creatorID,
		CreatedAt: &timestamp.Timestamp{Seconds: creationTime},
		OfferedBy: []*Entity{
			&Entity{
				Id:   cg.GetId(),
				Name: "Course Test Contract",
			},
		},
		EvidenceDocument: "use swarm hash here",
		DocumentPresence: "Physical",
	}
}

func generateFakeCoursesGrade(coursesIDS []string) (courses []*CourseGrade) {
	for _, courseID := range coursesIDS {
		assignments := generateFakeAssignmentGrades(rand.Intn(3))
		credentials := generateFakeAssignmentGradeCredentials(courseID, assignments)
		c := NewFakeCourseGrade(courseID, credentials)
		courses = append(courses, c)
	}
	return courses
}

func generateFakeCoursesGradeCredentials(courses []*CourseGrade) (credentials []*CourseGradeCredential) {
	for _, c := range courses {
		teacherID := c.Evaluators[0].GetId()
		c := NewFakeCourseGradeCredential(teacherID, c)
		credentials = append(credentials, c)
	}
	return credentials
}

// TODO: DRY
func NewFakeDiploma(facultyID string, credentials []*CourseGradeCredential) *Diploma {
	firstTimestamp := credentials[0].CreatedAt
	lastTimestamp := credentials[len(credentials)-1].CreatedAt
	var d duration.Duration
	d.Seconds = lastTimestamp.Seconds - firstTimestamp.Seconds
	d.Nanos = lastTimestamp.Nanos - firstTimestamp.Nanos
	diploma := &Diploma{
		Id:            facultyID,
		Name:          fmt.Sprintf("%s-%s", "Faculty", facultyID),
		Code:          fmt.Sprintf("%s%s", "F", facultyID[:4]),
		Category:      "BachelorDiploma",
		Type:          []string{"Diploma"},
		Language:      "en",
		Description:   "A bachelor diploma example",
		Duration:      &d,
		Supervisors:   credentials[0].Course.GetTeachers(),
		Evaluators:    credentials[0].Course.GetTeachers(),
		Student:       credentials[0].Course.GetStudent(),
		GradingSystem: "ECTS",
		ModeOfStudy:   "Full-time",
		TotalCredits:  180,
		Grades: map[string]int64{
			"gpa": int64(rand.Intn(100)), // TODO: compute base on courses' grades
		},
		StudentPresence: "Physical",
		Transcripts:     credentials, //TODO: store only map with ids and grades
	}
	return diploma
}

func NewFakeDiplomaCredential(creatorID string, d *Diploma) *DiplomaCredential {
	creationTime := time.Now().UTC().UnixNano()
	return &DiplomaCredential{
		Diploma:   d,
		CreatedBy: creatorID,
		CreatedAt: &timestamp.Timestamp{Seconds: creationTime},
		OfferedBy: []*Entity{
			&Entity{
				Id:   d.GetId(),
				Name: "Faculty Test Contract",
			},
		},
		EvidenceDocument: "use swarm hash here",
		DocumentPresence: "Physical",
	}
}

func generateFakeDiploma(facultyID string, coursesIDS []string) (diploma *Diploma) {
	courses := generateFakeCoursesGrade(coursesIDS)
	credentials := generateFakeCoursesGradeCredentials(courses)
	return NewFakeDiploma(facultyID, credentials)
}
