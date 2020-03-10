package course

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/r0qs/bbchain-dapp/src/core/backends"
	"github.com/r0qs/bbchain-dapp/src/core/course/contract"
	"github.com/stretchr/testify/assert"
)

var (
	teacherKey, teacherAddress     = backends.TestAccounts[0].Key, backends.TestAccounts[0].Address
	evaluatorKey, evaluatorAddress = backends.TestAccounts[1].Key, backends.TestAccounts[1].Address
	studentKey, studentAddress     = backends.TestAccounts[2].Key, backends.TestAccounts[2].Address
)

func deploy(backend *backends.TestBackend, prvKey *ecdsa.PrivateKey, owners []common.Address, quorum *big.Int) (common.Address, *contract.Course, error) {
	transactOpts := bind.NewKeyedTransactor(prvKey)
	startingTime, endingTime := backend.GetPeriod(uint64(100))
	courseAddr, _, course, err := contract.DeployCourse(transactOpts, backend, owners, quorum, startingTime, endingTime)
	if err != nil {
		return common.Address{}, nil, err
	}
	backend.Commit()
	return courseAddr, course, nil
}

func TestCourse(t *testing.T) {
	backend := backends.NewTestBackend()
	defer backend.Close()
	courseAddr, _, err := deploy(backend, teacherKey, []common.Address{teacherAddress, evaluatorAddress}, big.NewInt(2))
	if err != nil {
		t.Fatalf("deploy contract: expected no error, got %v", err)
	}

	course, err := NewCourse(courseAddr, backend)
	if err != nil {
		t.Fatalf("create contract: expected no error, got %v", err)
	}

	assert.Equal(t, course.Address(), courseAddr)
	if ok, err := course.IsOwner(&bind.CallOpts{Pending: true}, teacherAddress); !ok {
		t.Fatalf("IsOwner expected to be true but return: %t, %v", ok, err)
	}
}

func TestAddStudent(t *testing.T) {
	backend := backends.NewTestBackend()
	defer backend.Close()
	courseAddr, _, err := deploy(backend, teacherKey, []common.Address{teacherAddress, evaluatorAddress}, big.NewInt(2))
	if err != nil {
		t.Fatalf("deploy contract: expected no error, got %v", err)
	}
	course, err := NewCourse(courseAddr, backend)
	if err != nil {
		t.Fatalf("new contract: expected no error, got %v", err)
	}

	header, err := backend.HeaderByNumber(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	if ok, err := course.contract.IsStarted(nil); err != nil || !ok {
		t.Fatalf("Course should not be started yet in block: %v, but IsStarted returned: %t with error: %v", header.Number, ok, err)
	}

	// Increases block timestamp by 10 seconds
	err = backend.AdjustTime(10 * time.Second)
	if err != nil {
		t.Error(err)
	}

	// Add a student
	opts, _ := backend.GetTxOpts(teacherKey)
	if _, err := course.AddStudent(opts, studentAddress); err != nil {
		t.Fatalf("AddStudent expected to add a student but return: %v", err)
	}
	backend.Commit()

	// Verify if a student was added
	if ok, err := course.EnrolledStudents(&bind.CallOpts{Pending: true}, studentAddress); err != nil || !ok {
		t.Fatalf("EnrolledStudents expected student %v to be enrolled but return: %t, %v", studentAddress.Hex(), ok, err)
	}

	if ok, err := course.IsEnrolled(&bind.CallOpts{Pending: true}, studentAddress); err != nil || !ok {
		t.Fatalf("IsEnrolled expected student %v to be enrolled but return: %t, %v", studentAddress.Hex(), ok, err)
	}
}

func TestRemoveStudent(t *testing.T) {
	backend := backends.NewTestBackend()
	defer backend.Close()
	courseAddr, _, err := deploy(backend, teacherKey, []common.Address{teacherAddress, evaluatorAddress}, big.NewInt(2))
	if err != nil {
		t.Fatalf("deploy contract: expected no error, got %v", err)
	}
	course, _ := NewCourse(courseAddr, backend)

	// Add a student
	opts, _ := backend.GetTxOpts(teacherKey)
	course.AddStudent(opts, studentAddress)
	backend.Commit()

	// Remove a student
	if _, err := course.RemoveStudent(opts, studentAddress); err != nil {
		t.Fatalf("RemoveStudent expected to remove a student but return: %v", err)
	}
	backend.Commit()

	// Verify if a student was removed
	if ok, err := course.EnrolledStudents(&bind.CallOpts{Pending: true}, studentAddress); err != nil || ok {
		t.Fatalf("EnrolledStudents expected student %v to NOT be enrolled but return: %t, %v", studentAddress.Hex(), ok, err)
	}

	if ok, err := course.IsEnrolled(&bind.CallOpts{Pending: true}, studentAddress); err != nil || ok {
		t.Fatalf("IsEnrolled expected student %v to NOT be enrolled but return: %t, %v", studentAddress.Hex(), ok, err)
	}
}

func TestRenounceCourse(t *testing.T) {
	backend := backends.NewTestBackend()
	defer backend.Close()
	courseAddr, _, err := deploy(backend, teacherKey, []common.Address{teacherAddress, evaluatorAddress}, big.NewInt(2))
	if err != nil {
		t.Fatalf("deploy contract: expected no error, got %v", err)
	}
	course, _ := NewCourse(courseAddr, backend)

	// Add a student
	opts, _ := backend.GetTxOpts(teacherKey)
	course.AddStudent(opts, studentAddress)
	backend.Commit()

	opts, _ = backend.GetTxOpts(studentKey)
	if _, err := course.RenounceCourse(opts); err != nil {
		t.Fatalf("RenounceCourse expected to remove the sender (student) but return: %v", err)
	}
	backend.Commit()

	// Verify if a student was removed
	if ok, err := course.EnrolledStudents(&bind.CallOpts{Pending: true}, studentAddress); err != nil || ok {
		t.Fatalf("EnrolledStudents expected student %v to NOT be enrolled but return: %t, %v", studentAddress.Hex(), ok, err)
	}

	if ok, err := course.IsEnrolled(&bind.CallOpts{Pending: true}, studentAddress); err != nil || ok {
		t.Fatalf("IsEnrolled expected student %v to NOT be enrolled but return: %t, %v", studentAddress.Hex(), ok, err)
	}
}
