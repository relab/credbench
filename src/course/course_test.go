package course

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"log"
	"math/big"
	"testing"
)

var (
	teacherKey, teacherAddress     = getKeys("4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d")
	evaluatorKey, evaluatorAddress = getKeys("6cbed15c793ce57650b9877cf6fa156fbef513c4e6134f022a85b1ffdd59b2a1")
	studentKey, studentAddress     = getKeys("6370fd033278c143179d81c5526140625662b8daa446c22ee2d73db3707e620c")
)

func getKeys(hexkey string) (*ecdsa.PrivateKey, common.Address) {
	key, err := crypto.HexToECDSA(hexkey)
	if err != nil {
		log.Fatal(err)
	}
	address := crypto.PubkeyToAddress(key.PublicKey)
	return key, address
}

func newTestBackend() *backends.SimulatedBackend {
	return backends.NewSimulatedBackend(core.GenesisAlloc{
		teacherAddress:   {Balance: big.NewInt(1000000000)},
		evaluatorAddress: {Balance: big.NewInt(1000000000)},
		studentAddress:   {Balance: big.NewInt(1000000000)},
	}, 10000000)
}

func deploy(backend *backends.SimulatedBackend, prvKey *ecdsa.PrivateKey, owners []common.Address, quorum *big.Int) (common.Address, error) {
	transactOpts := bind.NewKeyedTransactor(prvKey)
	courseAddr, err := DeployCourse(transactOpts, backend, owners, quorum)
	if err != nil {
		return common.Address{}, err
	}
	backend.Commit()
	return courseAddr, nil
}

func newCourse(backend bind.ContractBackend, contractAddr common.Address, prvKey *ecdsa.PrivateKey) (*Course, error) {
	transactOpts := bind.NewKeyedTransactor(prvKey)
	return NewCourse(transactOpts, backend, contractAddr, prvKey)
}
func TestCourse(t *testing.T) {
	backend := newTestBackend()
	courseAddr, err := deploy(backend, teacherKey, []common.Address{teacherAddress, evaluatorAddress}, big.NewInt(2))
	if err != nil {
		t.Fatalf("deploy contract: expected no error, got %v", err)
	}

	course, err := newCourse(backend, courseAddr, teacherKey)
	if err != nil {
		t.Fatalf("create contract: expected no error, got %v", err)
	}

	assert.Equal(t, course.contractAddr, courseAddr)
	if ok, err := course.IsOwner(teacherAddress); !ok {
		t.Fatalf("IsOwner expected to be true but return: %t, %v", ok, err)
	}
}

func TestAddStudent(t *testing.T) {
	backend := newTestBackend()
	courseAddr, _ := deploy(backend, teacherKey, []common.Address{teacherAddress, evaluatorAddress}, big.NewInt(2))
	course, _ := newCourse(backend, courseAddr, teacherKey)

	// Add a student
	if _, err := course.AddStudent(studentAddress); err != nil {
		t.Fatalf("AddStudent expected to add a student but return: %v", err)
	}
	backend.Commit()

	// Verify if a student was added
	if ok, err := course.EnrolledStudents(studentAddress); err != nil || !ok {
		t.Fatalf("EnrolledStudents expected student %v to be enrolled but return: %t, %v", studentAddress.Hex(), ok, err)
	}

	if ok, err := course.IsEnrolled(studentAddress); err != nil || !ok {
		t.Fatalf("IsEnrolled expected student %v to be enrolled but return: %t, %v", studentAddress.Hex(), ok, err)
	}
}

func TestRemoveStudent(t *testing.T) {
	backend := newTestBackend()
	courseAddr, _ := deploy(backend, teacherKey, []common.Address{teacherAddress, evaluatorAddress}, big.NewInt(2))
	course, _ := newCourse(backend, courseAddr, teacherKey)
	course.AddStudent(studentAddress)
	backend.Commit()

	// Remove a student
	if _, err := course.RemoveStudent(studentAddress); err != nil {
		t.Fatalf("RemoveStudent expected to remove a student but return: %v", err)
	}
	backend.Commit()

	// Verify if a student was removed
	if ok, err := course.EnrolledStudents(studentAddress); err != nil || ok {
		t.Fatalf("EnrolledStudents expected student %v to NOT be enrolled but return: %t, %v", studentAddress.Hex(), ok, err)
	}

	if ok, err := course.IsEnrolled(studentAddress); err != nil || ok {
		t.Fatalf("IsEnrolled expected student %v to NOT be enrolled but return: %t, %v", studentAddress.Hex(), ok, err)
	}
}

func TestRenounceCourse(t *testing.T) {
	backend := newTestBackend()
	courseAddr, _ := deploy(backend, teacherKey, []common.Address{teacherAddress, evaluatorAddress}, big.NewInt(2))
	course, _ := newCourse(backend, courseAddr, teacherKey)
	course.AddStudent(studentAddress)
	backend.Commit()

	// Renouncing the course
	c, err := newCourse(backend, courseAddr, studentKey)
	if err != nil {
		t.Fatalf("NewCourse expected no error, got: %v", err)
	}

	if _, err := c.RenounceCourse(); err != nil {
		t.Fatalf("RenounceCourse expected to remove the sender (student) but return: %v", err)
	}
	backend.Commit()

	// Verify if a student was removed
	if ok, err := course.EnrolledStudents(studentAddress); err != nil || ok {
		t.Fatalf("EnrolledStudents expected student %v to NOT be enrolled but return: %t, %v", studentAddress.Hex(), ok, err)
	}

	if ok, err := course.IsEnrolled(studentAddress); err != nil || ok {
		t.Fatalf("IsEnrolled expected student %v to NOT be enrolled but return: %t, %v", studentAddress.Hex(), ok, err)
	}
}
