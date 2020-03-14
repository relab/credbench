package course

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"fmt"
	proto "github.com/golang/protobuf/proto"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/r0qs/bbchain-dapp/src/core/backends"
	"github.com/r0qs/bbchain-dapp/src/core/course/contract"
	pb "github.com/r0qs/bbchain-dapp/src/schemes"
	"github.com/stretchr/testify/assert"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

type TestCourse struct {
	Backend  *backends.TestBackend
	Owners   []backends.Account
	Students []backends.Account
	Course   *Course
}

func NewTestCourse(t *testing.T, owners backends.Accounts, quorum *big.Int) *TestCourse {
	backend := backends.NewTestBackend()

	courseAddr, _, err := deploy(backend, owners[0].Key, owners.Addresses(), big.NewInt(int64(len(owners))))
	if err != nil {
		t.Fatalf("deploy contract: expected no error, got %v", err)
	}
	course, err := NewCourse(courseAddr, backend)
	if err != nil {
		t.Fatalf("create contract: expected no error, got %v", err)
	}

	return &TestCourse{
		Backend: backend,
		Owners:  owners,
		Course:  course,
	}
}

func (tc *TestCourse) AddStudents(t *testing.T, students backends.Accounts) {
	opts, _ := tc.Backend.GetTxOpts(tc.Owners[0].Key)
	for _, addr := range students.Addresses() {
		_, err := tc.Course.AddStudent(opts, addr)
		if err != nil {
			t.Fatalf("AddStudent expected no error, got: %v", err)
		}
		tc.Backend.Commit()
	}
	tc.Students = students
}

func (tc *TestCourse) RegisterTestCredential(t *testing.T, to common.Address) [32]byte {
	digest := createTestDigest(tc.Owners[0].Address, to)
	opts, _ := tc.Backend.GetTxOpts(tc.Owners[0].Key)
	_, err := tc.Course.RegisterCredential(opts, to, digest)
	if err != nil {
		t.Fatalf("RegisterCredential expected no error, got: %v", err)
	}
	tc.Backend.Commit()
	proof := tc.Course.IssuedCredentials(nil, digest)
	assert.Equal(t, digest, proof.Digest)

	return digest
}

func (tc *TestCourse) ConfirmTestCredential(t *testing.T, from *ecdsa.PrivateKey, digest [32]byte) {
	opts, _ := tc.Backend.GetTxOpts(from)
	_, err := tc.Course.ConfirmCredential(opts, digest)
	if err != nil {
		t.Fatalf("ConfirmCredential expected no error, got: %v", err)
	}
	tc.Backend.Commit()
}

func deploy(backend *backends.TestBackend, prvKey *ecdsa.PrivateKey, owners []common.Address, quorum *big.Int) (common.Address, *contract.Course, error) {
	opts := bind.NewKeyedTransactor(prvKey)
	startingTime, endingTime := backend.GetPeriod(uint64(100))
	courseAddr, _, course, err := contract.DeployCourse(opts, backend, owners, quorum, startingTime, endingTime)
	if err != nil {
		return common.Address{}, nil, err
	}
	backend.Commit()
	return courseAddr, course, nil
}

func createTestDigest(teacherAddress, studentAddress common.Address) [32]byte {
	c := rand.Intn(100)
	assignment := &pb.Assignment{
		Id:          hashString(fmt.Sprintf("%s%d", "AssignmentFile-", c)),
		Name:        fmt.Sprintf("%s%d", "Exam ", c),
		Code:        fmt.Sprintf("%s%d", "EX-", c),
		Category:    "InternalActivity",
		Type:        []string{"MandatoryActivity"},
		Language:    "en",
		Description: "This is an exam description",
		Evaluators: []*pb.Evaluator{
			&pb.Evaluator{
				Id:   fmt.Sprintf("%x", teacherAddress),
				Name: "Teacher Name",
				Role: "teacher",
			},
		},
		Student: &pb.Student{
			Id:   fmt.Sprintf("%x", studentAddress),
			Name: "Student Name",
		},
		Grade:           42,
		SubjectPresence: "Physical",
	}
	data, _ := proto.Marshal(assignment)
	return sha256.Sum256(data)
}

func hashString(s string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
}

// Start tests

func TestNewCourse(t *testing.T) {
	tc := NewTestCourse(t, backends.TestAccounts[:2], big.NewInt(2))
	defer tc.Backend.Close()

	if ok, err := tc.Course.IsOwner(&bind.CallOpts{Pending: true}, tc.Owners[0].Address); !ok {
		t.Fatalf("IsOwner expected to be true but return: %t, %v", ok, err)
	}
}

func TestAddStudent(t *testing.T) {
	tc := NewTestCourse(t, backends.TestAccounts[:1], big.NewInt(1))
	defer tc.Backend.Close()

	header, err := tc.Backend.HeaderByNumber(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	if ok, err := tc.Course.contract.IsStarted(nil); err != nil || !ok {
		t.Fatalf("Course should not be started yet in block: %v, but IsStarted returned: %t with error: %v", header.Number, ok, err)
	}

	// Increases block timestamp by 10 seconds
	err = tc.Backend.AdjustTime(10 * time.Second)
	if err != nil {
		t.Error(err)
	}

	// Add a student
	studentAddress := backends.TestAccounts[2].Address
	opts, _ := tc.Backend.GetTxOpts(tc.Owners[0].Key)
	if _, err := tc.Course.AddStudent(opts, studentAddress); err != nil {
		t.Fatalf("AddStudent expected to add a student but return: %v", err)
	}
	tc.Backend.Commit()

	// Verify if a student was added
	if ok, err := tc.Course.EnrolledStudents(&bind.CallOpts{Pending: true}, studentAddress); err != nil || !ok {
		t.Fatalf("EnrolledStudents expected student %v to be enrolled but return: %t, %v", studentAddress.Hex(), ok, err)
	}

	if ok, err := tc.Course.IsEnrolled(&bind.CallOpts{Pending: true}, studentAddress); err != nil || !ok {
		t.Fatalf("IsEnrolled expected student %v to be enrolled but return: %t, %v", studentAddress.Hex(), ok, err)
	}
}

func TestRemoveStudent(t *testing.T) {
	tc := NewTestCourse(t, backends.TestAccounts[:1], big.NewInt(1))
	defer tc.Backend.Close()

	student := backends.TestAccounts[2]
	studentAddress := student.Address
	tc.AddStudents(t, backends.Accounts{student})

	// Remove a student
	opts, _ := tc.Backend.GetTxOpts(tc.Owners[0].Key)
	if _, err := tc.Course.RemoveStudent(opts, studentAddress); err != nil {
		t.Fatalf("RemoveStudent expected to remove a student but return: %v", err)
	}
	tc.Backend.Commit()

	// Verify if a student was removed
	if ok, err := tc.Course.EnrolledStudents(&bind.CallOpts{Pending: true}, studentAddress); err != nil || ok {
		t.Fatalf("EnrolledStudents expected student %v to NOT be enrolled but return: %t, %v", studentAddress.Hex(), ok, err)
	}

	if ok, err := tc.Course.IsEnrolled(&bind.CallOpts{Pending: true}, studentAddress); err != nil || ok {
		t.Fatalf("IsEnrolled expected student %v to NOT be enrolled but return: %t, %v", studentAddress.Hex(), ok, err)
	}
}

func TestRenounceCourse(t *testing.T) {
	tc := NewTestCourse(t, backends.TestAccounts[:1], big.NewInt(1))
	defer tc.Backend.Close()

	student := backends.TestAccounts[2]
	studentKey := student.Key
	studentAddress := student.Address
	tc.AddStudents(t, backends.Accounts{student})

	opts, _ := tc.Backend.GetTxOpts(studentKey)
	if _, err := tc.Course.RenounceCourse(opts); err != nil {
		t.Fatalf("RenounceCourse expected to remove the sender (student) but return: %v", err)
	}
	tc.Backend.Commit()

	// Verify if a student was removed
	if ok, err := tc.Course.EnrolledStudents(&bind.CallOpts{Pending: true}, studentAddress); err != nil || ok {
		t.Fatalf("EnrolledStudents expected student %v to NOT be enrolled but return: %t, %v", studentAddress.Hex(), ok, err)
	}

	if ok, err := tc.Course.IsEnrolled(&bind.CallOpts{Pending: true}, studentAddress); err != nil || ok {
		t.Fatalf("IsEnrolled expected student %v to NOT be enrolled but return: %t, %v", studentAddress.Hex(), ok, err)
	}
}

func TestIssuerRegisterCredential(t *testing.T) {
	tc := NewTestCourse(t, backends.TestAccounts[:1], big.NewInt(1))
	defer tc.Backend.Close()

	student := backends.TestAccounts[2]
	studentAddress := student.Address
	tc.AddStudents(t, backends.Accounts{student})

	digest := tc.RegisterTestCredential(t, studentAddress)

	proof := tc.Course.IssuedCredentials(nil, digest)

	assert.Equal(t, studentAddress, proof.Subject, "Subject address should be equal")
	assert.Equal(t, tc.Owners[0].Address, proof.Issuer, "Subject address should be equal")
	assert.Equal(t, digest, proof.Digest, "Assignment digest should be equal")
	assert.False(t, proof.SubjectSigned)
}

func TestStudentSignCredential(t *testing.T) {
	tc := NewTestCourse(t, backends.TestAccounts[:1], big.NewInt(1))
	defer tc.Backend.Close()

	student := backends.TestAccounts[2]
	studentKey := student.Key
	studentAddress := student.Address
	tc.AddStudents(t, backends.Accounts{student})

	digest := tc.RegisterTestCredential(t, studentAddress)

	proof := tc.Course.IssuedCredentials(nil, digest)
	assert.False(t, proof.SubjectSigned)

	tc.ConfirmTestCredential(t, studentKey, digest)

	proof = tc.Course.IssuedCredentials(nil, digest)
	assert.True(t, proof.SubjectSigned)
}

func TestIssuerAggregateCredential(t *testing.T) {
	tc := NewTestCourse(t, backends.TestAccounts[:1], big.NewInt(1))
	defer tc.Backend.Close()

	student := backends.TestAccounts[2]
	studentKey := student.Key
	studentAddress := student.Address
	tc.AddStudents(t, backends.Accounts{student})

	var digests [][32]byte
	for i := 0; i < 2; i++ {
		d := tc.RegisterTestCredential(t, studentAddress)
		digests = append(digests, d)
		tc.ConfirmTestCredential(t, studentKey, d)
	}

	// Force end of the course
	endingTime, _ := tc.Course.contract.EndingTime(nil)
	err := tc.Backend.IncreaseTime(time.Duration(endingTime.Int64()) * time.Second)
	if err != nil {
		t.Error(err)
	}

	ended, err := tc.Course.contract.HasEnded(nil)
	if err != nil {
		t.Error(err)
	}
	assert.True(t, ended)

	opts, _ := tc.Backend.GetTxOpts(tc.Owners[0].Key)
	_, err = tc.Course.contract.AggregateCredentials(opts, studentAddress)
	if err != nil {
		t.Fatalf("AggregateCredentials expected no error, got: %v", err)
	}
	tc.Backend.Commit()

	aggregatedDigest, err := tc.Course.contract.GetProof(nil, studentAddress)
	if err != nil {
		t.Error(err)
	}

	expectedDigests, err := backends.EncodeByteArray(digests)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, expectedDigests, aggregatedDigest)
}
