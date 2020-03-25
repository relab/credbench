package course

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/relab/bbchain-dapp/src/core/accounts"
	"github.com/relab/bbchain-dapp/src/core/backends"
	"github.com/relab/bbchain-dapp/src/core/course/contract"
	"github.com/stretchr/testify/assert"

	pb "github.com/relab/bbchain-dapp/src/schemes"
)

type TestCourse struct {
	Backend    *backends.TestBackend
	Evaluators []backends.Account
	Students   []backends.Account
	Course     *Course
}

func NewTestCourse(t *testing.T, evaluators backends.Accounts, quorum *big.Int) *TestCourse {
	backend := backends.NewTestBackend()

	courseAddr, _, err := deploy(backend, evaluators[0].Key, evaluators.Addresses(), big.NewInt(int64(len(evaluators))))
	if err != nil {
		t.Fatalf("deploy contract: expected no error, got %v", err)
	}
	course, err := NewCourse(courseAddr, backend)
	if err != nil {
		t.Fatalf("create contract: expected no error, got %v", err)
	}

	return &TestCourse{
		Backend:    backend,
		Evaluators: evaluators,
		Course:     course,
	}
}

func (tc *TestCourse) AddStudents(t *testing.T, students backends.Accounts) {
	opts, _ := tc.Backend.GetTxOpts(tc.Evaluators[0].Key)
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
	evaluatorAddr := tc.Evaluators[0].Address.Hex()
	courseEntity := &pb.Entity{
		Id:   tc.Course.Address().Hex(),
		Name: "Course Test Contract",
	}
	ag := pb.NewFakeAssignmentGrade(evaluatorAddr, to.Hex())
	credential := pb.NewFakeAssignmentGradeCredential(evaluatorAddr, courseEntity, ag)
	digest := pb.Hash(credential)

	ch := make(chan *contract.CourseCredentialIssued)
	sub, _ := tc.Course.contract.WatchCredentialIssued(nil, ch, [][32]byte{digest}, []common.Address{to}, nil)
	defer func() {
		sub.Unsubscribe()
	}()

	opts, _ := tc.Backend.GetTxOpts(tc.Evaluators[0].Key)
	_, err := tc.Course.RegisterCredential(opts, to, digest)
	if err != nil {
		t.Fatalf("RegisterCredential expected no error, got: %v", err)
	}
	tc.Backend.Commit()

	event := <-ch
	assert.Equal(t, digest, event.Digest)

	proof := tc.Course.IssuedCredentials(nil, digest)
	assert.Equal(t, digest, proof.Digest)

	return digest
}

func (tc *TestCourse) ConfirmTestCredential(t *testing.T, from *ecdsa.PrivateKey, digest [32]byte) {
	ch := make(chan *contract.CourseCredentialSigned)
	sub, _ := tc.Course.contract.WatchCredentialSigned(nil, ch, nil, [][32]byte{digest})
	defer func() {
		sub.Unsubscribe()
	}()

	opts, _ := tc.Backend.GetTxOpts(from)
	_, err := tc.Course.ConfirmCredential(opts, digest)
	if err != nil {
		t.Fatalf("ConfirmCredential expected no error, got: %v", err)
	}
	tc.Backend.Commit()

	event := <-ch
	assert.Equal(t, accounts.GetAddress(from), event.Signer)
}

func deploy(backend *backends.TestBackend, prvKey *ecdsa.PrivateKey, evaluators []common.Address, quorum *big.Int) (common.Address, *contract.Course, error) {
	opts := bind.NewKeyedTransactor(prvKey)
	startingTime, endingTime := backend.GetPeriod(uint64(100))
	courseAddr, _, course, err := contract.DeployCourse(opts, backend, evaluators, quorum, startingTime, endingTime)
	if err != nil {
		return common.Address{}, nil, err
	}
	backend.Commit()
	return courseAddr, course, nil
}

// Start tests

func TestNewCourse(t *testing.T) {
	evaluatorsAccount := backends.TestAccounts[:2]
	tc := NewTestCourse(t, evaluatorsAccount, big.NewInt(2))
	defer tc.Backend.Close()

	// Calling Owners contract methods
	if ok, err := tc.Course.IsOwner(&bind.CallOpts{Pending: true}, tc.Evaluators[0].Address); !ok {
		t.Fatalf("IsOwner expected to be true but return: %t, %v", ok, err)
	}

	owners, err := tc.Course.OwnersList(&bind.CallOpts{Pending: true})
	if err != nil {
		t.Fatalf("OwnersList expected no errors but got: %v", err)
	}
	assert.ElementsMatch(t, evaluatorsAccount.Addresses(), owners)
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
	opts, _ := tc.Backend.GetTxOpts(tc.Evaluators[0].Key)
	if _, err := tc.Course.AddStudent(opts, studentAddress); err != nil {
		t.Fatalf("AddStudent expected to add a student but return: %v", err)
	}
	tc.Backend.Commit()

	for iter, _ := tc.Course.contract.FilterStudentAdded(nil, []common.Address{studentAddress}, nil); iter.Next(); {
		event := iter.Event
		assert.Equal(t, studentAddress, event.Student)
	}

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
	opts, _ := tc.Backend.GetTxOpts(tc.Evaluators[0].Key)
	if _, err := tc.Course.RemoveStudent(opts, studentAddress); err != nil {
		t.Fatalf("RemoveStudent expected to remove a student but return: %v", err)
	}
	tc.Backend.Commit()

	for iter, _ := tc.Course.contract.FilterStudentRemoved(nil, []common.Address{studentAddress}, nil); iter.Next(); {
		event := iter.Event
		assert.Equal(t, studentAddress, event.Student)
	}

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

	for iter, _ := tc.Course.contract.FilterStudentRemoved(nil, []common.Address{studentAddress}, nil); iter.Next(); {
		event := iter.Event
		assert.Equal(t, studentAddress, event.Student)
	}

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
	assert.Equal(t, tc.Evaluators[0].Address, proof.Issuer, "Subject address should be equal")
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
	endingTime, _ := tc.Course.EndingTime(nil)
	err := tc.Backend.IncreaseTime(time.Duration(endingTime.Int64()) * time.Second)
	if err != nil {
		t.Error(err)
	}

	ended, err := tc.Course.HasEnded(nil)
	if err != nil {
		t.Error(err)
	}
	assert.True(t, ended)

	opts, _ := tc.Backend.GetTxOpts(tc.Evaluators[0].Key)
	_, err = tc.Course.AggregateCredentials(opts, studentAddress)
	if err != nil {
		t.Fatalf("AggregateCredentials expected no error, got: %v", err)
	}
	tc.Backend.Commit()

	aggregatedDigest, err := tc.Course.GetProof(nil, studentAddress)
	if err != nil {
		t.Error(err)
	}

	expectedDigest, err := backends.EncodeByteArray(digests)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, expectedDigest, aggregatedDigest)
}

func TestVerifyCredential(t *testing.T) {
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

	endingTime, _ := tc.Course.EndingTime(nil)
	err := tc.Backend.IncreaseTime(time.Duration(endingTime.Int64()) * time.Second)
	if err != nil {
		t.Error(err)
	}

	ended, err := tc.Course.HasEnded(nil)
	if err != nil {
		t.Error(err)
	}
	assert.True(t, ended)

	opts, _ := tc.Backend.GetTxOpts(tc.Evaluators[0].Key)
	_, err = tc.Course.AggregateCredentials(opts, studentAddress)
	if err != nil {
		t.Fatalf("AggregateCredentials expected no error, got: %v", err)
	}
	tc.Backend.Commit()

	expectedDigest, err := backends.EncodeByteArray(digests)
	if err != nil {
		t.Error(err)
	}

	b, err := tc.Course.VerifyCredential(nil, studentAddress, expectedDigest)
	if err != nil {
		t.Error(err)
	}

	assert.True(t, b)
}
