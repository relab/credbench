package schemes

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"time"

	proto "github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/timestamp"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func hashString(s string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
}

func NewFakeAssignmentGrade(teacherId, studentId string) *AssignmentGrade {
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
				Id: teacherId,
			},
		},
		Student: &Entity{
			Id: studentId,
		},
		Grade:           42,
		StudentPresence: "Physical",
	}
}

func (ag *AssignmentGrade) Hash() [32]byte {
	data, _ := proto.Marshal(ag)
	return sha256.Sum256(data)
}

func NewFakeAssignmentGradeCredential(creatorId string, courseEntity *Entity, ag *AssignmentGrade) *AssignmentGradeCredential {
	creationTime := time.Now().UTC().UnixNano()
	return &AssignmentGradeCredential{
		Assignment:       ag,
		CreatedBy:        creatorId,
		CreatedAt:        &timestamp.Timestamp{Seconds: creationTime},
		OfferedBy:        []*Entity{courseEntity},
		EvidenceDocument: "",
		DocumentPresence: "Physical",
	}
}

func (ag *AssignmentGradeCredential) Hash() [32]byte {
	data, _ := proto.Marshal(ag)
	return sha256.Sum256(data)
}
