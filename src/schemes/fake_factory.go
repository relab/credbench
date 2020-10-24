package schemes

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func hashString(s string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
}

func NewFakeAssignmentGrade(teacherID, studentID string) *AssignmentGrade {
	c := rand.Intn(10000) //FIXME: if the same number is chosen twice, two assignments will have the same hash in the test case and some tests will fail since isn't possible to issue two assignments with the same hash
	return &AssignmentGrade{
		Id:          hashString(fmt.Sprintf("%s%d", "AssignmentFile-", c)),
		Name:        fmt.Sprintf("%s%d", "Exam ", c),
		Code:        fmt.Sprintf("%s%d", "EX-", c),
		Category:    "InternalActivity",
		Type:        []string{"MandatoryActivity"},
		Language:    "en",
		Description: "This is an exam description",
		Evaluators: []*Entity{
			{
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
	creationTime := time.Now().Unix()
	return &AssignmentGradeCredential{
		Assignment:       ag,
		CreatedBy:        creatorID,
		CreatedAt:        &timestamppb.Timestamp{Seconds: creationTime},
		OfferedBy:        []*Entity{courseEntity},
		EvidenceDocument: "use swarm hash here",
		DocumentPresence: "Physical",
	}
}

func generateFakeAssignmentGrades(teacherID, studentID string, n int) (assignments []*AssignmentGrade) {
	for i := 0; i < n; i++ {
		ag := NewFakeAssignmentGrade(teacherID, studentID)
		assignments = append(assignments, ag)
	}
	return assignments
}

func generateFakeAssignmentGradeCredentials(creatorID string, courseID string, assignments []*AssignmentGrade) (credentials []*AssignmentGradeCredential) {
	courseEntity := &Entity{
		Id:   courseID,
		Name: "Course Test Contract",
	}
	for _, a := range assignments {
		c := NewFakeAssignmentGradeCredential(creatorID, courseEntity, a)
		credentials = append(credentials, c)
	}
	return credentials
}

// TODO: DRY
func NewFakeCourseGrade(courseID string, credentials []*AssignmentGradeCredential) *CourseGrade {
	firstTimestamp := credentials[0].CreatedAt
	lastTimestamp := credentials[len(credentials)-1].CreatedAt
	var d durationpb.Duration
	d.Seconds = lastTimestamp.Seconds - firstTimestamp.Seconds
	d.Nanos = lastTimestamp.Nanos - firstTimestamp.Nanos
	cgrade := &CourseGrade{
		Id:              courseID,
		Name:            fmt.Sprintf("%s-%s", "Course", courseID),
		Code:            fmt.Sprintf("%s%s", "C", courseID[2:6]),
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
	creationTime := time.Now().Unix()
	return &CourseGradeCredential{
		Course:    cg,
		CreatedBy: creatorID,
		CreatedAt: &timestamppb.Timestamp{Seconds: creationTime},
		OfferedBy: []*Entity{
			{
				Id:   cg.GetId(),
				Name: "Course Test Contract",
			},
		},
		EvidenceDocument: "use swarm hash here",
		DocumentPresence: "Physical",
	}
}

func GenerateFakeCoursesGrade(teacherID, studentID string, coursesIDS []string) (courses []*CourseGrade) {
	for _, courseID := range coursesIDS {
		assignments := generateFakeAssignmentGrades(teacherID, studentID, 4) // 4 assignments per course
		credentials := generateFakeAssignmentGradeCredentials(teacherID, courseID, assignments)
		c := NewFakeCourseGrade(courseID, credentials)
		courses = append(courses, c)
	}
	return courses
}

func GenerateFakeCoursesGradeCredentials(creatorID string, courses []*CourseGrade) (credentials []*CourseGradeCredential) {
	for _, c := range courses {
		c := NewFakeCourseGradeCredential(creatorID, c)
		credentials = append(credentials, c)
	}
	return credentials
}

// TODO: DRY
func NewFakeDiploma(facultyID string, credentials []*CourseGradeCredential) *Diploma {
	firstTimestamp := credentials[0].CreatedAt
	lastTimestamp := credentials[len(credentials)-1].CreatedAt
	var d durationpb.Duration
	d.Seconds = lastTimestamp.Seconds - firstTimestamp.Seconds
	d.Nanos = lastTimestamp.Nanos - firstTimestamp.Nanos
	diploma := &Diploma{
		Id:            facultyID,
		Name:          fmt.Sprintf("%s-%s", "Faculty", facultyID),
		Code:          fmt.Sprintf("%s%s", "F", facultyID[2:6]),
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
		Courses:         credentials, //TODO: store only map with ids and grades
	}
	return diploma
}

func NewFakeDiplomaCredential(creatorID string, d *Diploma) *DiplomaCredential {
	creationTime := time.Now().Unix()
	return &DiplomaCredential{
		Diploma:   d,
		CreatedBy: creatorID,
		CreatedAt: &timestamppb.Timestamp{Seconds: creationTime},
		OfferedBy: []*Entity{
			{
				Id:   d.GetId(),
				Name: "Faculty Test Contract",
			},
		},
		EvidenceDocument: "use swarm hash here",
		DocumentPresence: "Physical",
	}
}

func GenerateFakeDiploma(facultyID, teacherID, studentID string, coursesIDS []string) (diploma *Diploma) {
	courses := GenerateFakeCoursesGrade(teacherID, studentID, coursesIDS)
	credentials := GenerateFakeCoursesGradeCredentials(teacherID, courses)
	return NewFakeDiploma(facultyID, credentials)
}
