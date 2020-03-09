package course

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/r0qs/bbchain-dapp/src/core/course/contract"
	"github.com/stretchr/testify/assert"
)

var (
	teacherKey, teacherAddress     = getKeys("4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d")
	evaluatorKey, evaluatorAddress = getKeys("6cbed15c793ce57650b9877cf6fa156fbef513c4e6134f022a85b1ffdd59b2a1")
	studentKey, studentAddress     = getKeys("6370fd033278c143179d81c5526140625662b8daa446c22ee2d73db3707e620c")
)

// duration in seconds
func getPeriod(backend *backends.SimulatedBackend, duration uint64) (*big.Int, *big.Int) {
	header, _ := backend.HeaderByNumber(context.Background(), nil)
	// Every backend.Commit() increases the block time in 10 secs
	// so we calculate the start time to in the next block
	startingTime := header.Time + 10
	endingTime := startingTime + duration
	return new(big.Int).SetUint64(startingTime), new(big.Int).SetUint64(endingTime)
}

func getKeys(hexkey string) (*ecdsa.PrivateKey, common.Address) {
	key, err := crypto.HexToECDSA(hexkey)
	if err != nil {
		log.Fatal(err)
	}
	address := crypto.PubkeyToAddress(key.PublicKey)
	return key, address
}

func getTxOpts(backend *backends.SimulatedBackend, key *ecdsa.PrivateKey) (*bind.TransactOpts, error) {
	gasPrice, err := backend.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("Failed to estimate the gas price: %v", err)
	}
	opts := bind.NewKeyedTransactor(key)
	opts.GasLimit = uint64(6721975)
	opts.GasPrice = gasPrice
	return opts, nil
}

func newTestBackend() *backends.SimulatedBackend {
	return backends.NewSimulatedBackend(core.GenesisAlloc{
		teacherAddress:   {Balance: big.NewInt(1000000000)},
		evaluatorAddress: {Balance: big.NewInt(1000000000)},
		studentAddress:   {Balance: big.NewInt(1000000000)},
	}, 10000000)
}

func deploy(backend *backends.SimulatedBackend, prvKey *ecdsa.PrivateKey, owners []common.Address, quorum *big.Int) (common.Address, *contract.Course, error) {
	transactOpts := bind.NewKeyedTransactor(prvKey)
	startingTime, endingTime := getPeriod(backend, uint64(100))
	courseAddr, _, course, err := contract.DeployCourse(transactOpts, backend, owners, quorum, startingTime, endingTime)
	if err != nil {
		return common.Address{}, nil, err
	}
	backend.Commit() // every Commit increase the block time in 10 secs
	return courseAddr, course, nil
}

func TestCourse(t *testing.T) {
	backend := newTestBackend()
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
	backend := newTestBackend()
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
	opts, _ := getTxOpts(backend, teacherKey)
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
	backend := newTestBackend()
	defer backend.Close()
	courseAddr, _, err := deploy(backend, teacherKey, []common.Address{teacherAddress, evaluatorAddress}, big.NewInt(2))
	if err != nil {
		t.Fatalf("deploy contract: expected no error, got %v", err)
	}
	course, _ := NewCourse(courseAddr, backend)

	// Add a student
	opts, _ := getTxOpts(backend, teacherKey)
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
	backend := newTestBackend()
	defer backend.Close()
	courseAddr, _, err := deploy(backend, teacherKey, []common.Address{teacherAddress, evaluatorAddress}, big.NewInt(2))
	if err != nil {
		t.Fatalf("deploy contract: expected no error, got %v", err)
	}
	course, _ := NewCourse(courseAddr, backend)

	// Add a student
	opts, _ := getTxOpts(backend, teacherKey)
	course.AddStudent(opts, studentAddress)
	backend.Commit()

	opts, _ = getTxOpts(backend, studentKey)
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
