package faculty

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	// proto "github.com/golang/protobuf/proto"
	// pb "github.com/r0qs/bbchain-dapp/src/schemes"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/r0qs/bbchain-dapp/src/core/backends"
	"github.com/r0qs/bbchain-dapp/src/core/course"
	"github.com/r0qs/bbchain-dapp/src/core/faculty/contract"
	"github.com/stretchr/testify/assert"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

type TestFaculty struct {
	Backend    *backends.TestBackend
	Adms       []backends.Account
	Evaluators []backends.Account
	Students   []backends.Account
	Faculty    *Faculty
}

// if possible make these methods generic (reflect)
func NewTestFaculty(t *testing.T, adms backends.Accounts, quorum *big.Int) *TestFaculty {
	backend := backends.NewTestBackend()

	facultyAddr, _, err := deploy(backend, adms[0].Key, adms.Addresses(), big.NewInt(int64(len(adms))))
	if err != nil {
		t.Fatalf("deploy contract: expected no error, got %v", err)
	}
	faculty, err := NewFaculty(facultyAddr, backend)
	if err != nil {
		t.Fatalf("create contract: expected no error, got %v", err)
	}

	return &TestFaculty{
		Backend: backend,
		Adms:    adms,
		Faculty: faculty,
	}
}

// TODO: make generic (reflect)
func deploy(backend *backends.TestBackend, prvKey *ecdsa.PrivateKey, adms []common.Address, quorum *big.Int) (common.Address, *contract.Faculty, error) {
	opts := bind.NewKeyedTransactor(prvKey)
	facultyAddr, _, faculty, err := contract.DeployFaculty(opts, backend, adms, quorum)
	if err != nil {
		return common.Address{}, nil, err
	}
	backend.Commit()
	return facultyAddr, faculty, nil
}

// todo: create diploma
func TestNewFaculty(t *testing.T) {
	admsAccount := backends.TestAccounts[:2]
	tf := NewTestFaculty(t, admsAccount, big.NewInt(2))
	defer tf.Backend.Close()

	if ok, err := tf.Faculty.IsOwner(&bind.CallOpts{Pending: true}, tf.Adms[0].Address); !ok {
		t.Fatalf("IsOwner expected to be true but return: %t, %v", ok, err)
	}

	adms, err := tf.Faculty.OwnersList(&bind.CallOpts{Pending: true})
	if err != nil {
		t.Fatalf("OwnersList expected no errors but got: %v", err)
	}
	assert.ElementsMatch(t, admsAccount.Addresses(), adms)
}

func TestFacultyCreateCourse(t *testing.T) {
	adms := backends.TestAccounts[:2]
	evaluators := backends.TestAccounts[2:4]

	tf := NewTestFaculty(t, adms, big.NewInt(int64(len(adms))))
	defer tf.Backend.Close()

	if ok, err := tf.Faculty.IsOwner(&bind.CallOpts{Pending: true}, tf.Adms[0].Address); !ok {
		t.Fatalf("IsOwner expected to be true but return: %t, %v", ok, err)
	}

	startingTime, endingTime := tf.Backend.GetPeriod(uint64(100))
	semester := sha256.Sum256([]byte("spring2020"))
	opts := bind.NewKeyedTransactor(adms[0].Key)
	_, err := tf.Faculty.contract.CreateCourse(opts, semester, evaluators.Addresses(), big.NewInt(int64(len(evaluators))), startingTime, endingTime)
	if err != nil {
		t.Fatalf("CreateCourse expected no error but got: %v", err)
	}
	tf.Backend.Commit()

	block, _ := tf.Backend.BlockByNumber(context.Background(), nil)
	query := ethereum.FilterQuery{
		ToBlock: block.Number(),
		Addresses: []common.Address{
			tf.Faculty.Address(),
		},
	}

	logs, err := tf.Backend.FilterLogs(context.Background(), query)
	if err != nil {
		t.Fatal(err)
	}

	evIssuerAdded, err := tf.Faculty.contract.ParseIssuerAdded(logs[0])
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, evIssuerAdded.AddedBy, adms[0].Address)
	courseAddr := evIssuerAdded.IssuerAddress
	c, err := tf.Faculty.contract.Courses(nil, courseAddr)
	assert.Equal(t, courseAddr, c)

	evCourseCreated, err := tf.Faculty.contract.ParseCourseCreated(logs[1])
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, evCourseCreated.CreatedBy, adms[0].Address)
	assert.Equal(t, evCourseCreated.CourseAddress, c)
	assert.Equal(t, evCourseCreated.Semester, semester)
	assert.Equal(t, evCourseCreated.Quorum, big.NewInt(int64(len(evaluators))))
	assert.ElementsMatch(t, evCourseCreated.Teachers, evaluators.Addresses())

	courseInstance, err := course.NewCourse(courseAddr, tf.Backend)
	if err != nil {
		t.Fatalf("create contract: expected no error, got %v", err)
	}

	teachers, err := courseInstance.OwnersList(nil)
	if err != nil {
		t.Fatalf("OwnersList expected no errors but got: %v", err)
	}
	assert.ElementsMatch(t, evaluators.Addresses(), teachers)
}
