package faculty

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/relab/bbchain-dapp/src/core/accounts"
	"github.com/relab/bbchain-dapp/src/core/backends"
	"github.com/relab/bbchain-dapp/src/core/course"
	"github.com/relab/bbchain-dapp/src/core/encode"
	"github.com/relab/bbchain-dapp/src/core/faculty/contract"

	pb "github.com/relab/bbchain-dapp/src/schemes"
)

type TestFaculty struct {
	Backend    *backends.TestBackend
	Adms       []backends.Account
	Evaluators []backends.Account
	Students   []backends.Account
	Faculty    *Faculty
}

func NewTestFaculty(t *testing.T, adms backends.Accounts, quorum *big.Int) *TestFaculty {
	backend := backends.NewTestBackend()

	facultyAddr, _, err := deployFaculty(backend, adms[0].Key, adms.Addresses(), big.NewInt(int64(len(adms))))
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

func deployFaculty(backend *backends.TestBackend, prvKey *ecdsa.PrivateKey, adms []common.Address, quorum *big.Int) (common.Address, *contract.Faculty, error) {
	opts := bind.NewKeyedTransactor(prvKey)
	facultyAddr, _, faculty, err := contract.DeployFaculty(opts, backend, adms, quorum)
	if err != nil {
		return common.Address{}, nil, err
	}
	backend.Commit()
	return facultyAddr, faculty, nil
}

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

	semester := sha256.Sum256([]byte("spring2020"))
	opts := bind.NewKeyedTransactor(adms[0].Key)
	_, err := tf.Faculty.CreateCourse(opts, semester, evaluators.Addresses(), big.NewInt(int64(len(evaluators))))
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
	ok, err := tf.Faculty.IsIssuer(nil, courseAddr)
	assert.True(t, ok)

	evCourseCreated, err := tf.Faculty.contract.ParseCourseCreated(logs[1])
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, evCourseCreated.CreatedBy, adms[0].Address)
	assert.Equal(t, evCourseCreated.CourseAddress, courseAddr)
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

func TestCreateDiploma(t *testing.T) {
	adms := backends.TestAccounts[:2]
	evaluators := backends.TestAccounts[2:4]
	student := backends.TestAccounts[4]

	tf := NewTestFaculty(t, adms, big.NewInt(int64(len(adms))))
	defer tf.Backend.Close()

	var coursesAddresses []common.Address
	for i := 0; i < 4; i++ {
		// adm creates course
		semester := sha256.Sum256([]byte("spring2020"))
		opts, _ := accounts.GetTxOpts(adms[0].Key, tf.Backend)
		_, err := tf.Faculty.CreateCourse(opts, semester, evaluators.Addresses(), big.NewInt(int64(len(evaluators))))
		if err != nil {
			t.Fatalf("CreateCourse expected no error but got: %v", err)
		}
		tf.Backend.Commit()

		// get course instance
		courseAddr, _ := tf.Faculty.contract.Issuers(nil, big.NewInt(int64(i)))
		coursesAddresses = append(coursesAddresses, courseAddr)

		courseInstance, err := course.NewCourse(courseAddr, tf.Backend)
		if err != nil {
			t.Fatalf("create contract: expected no error, got %v", err)
		}
		if ok, _ := courseInstance.IsOwner(nil, evaluators[0].Address); !ok {
			t.Fatalf("Evaluator %v is expected to be owner of the course", evaluators[0].Address.Hex())
		}

		// teacher adds student
		opts, _ = accounts.GetTxOpts(evaluators[0].Key, tf.Backend)
		_, err = courseInstance.AddStudent(opts, student.Address)
		if err != nil {
			t.Fatalf("AddStudent expected no error, got: %v", err)
		}
		tf.Backend.Commit()

		if ok, _ := courseInstance.IsEnrolled(nil, student.Address); !ok {
			t.Fatalf("Student %v is expected to be enrolled in the course", student.Address.Hex())
		}
	}

	cAddresses := make([]string, len(coursesAddresses))
	for i, addr := range coursesAddresses {
		cAddresses[i] = addr.Hex()
	}

	// TODO pass list of teachers and students
	diploma := pb.GenerateFakeDiploma(tf.Faculty.Address().Hex(), evaluators[0].Address.Hex(), student.Address.Hex(), cAddresses)

	var expectedDigests [][32]byte
	for _, c := range diploma.Courses {
		caddr := common.HexToAddress(c.Course.GetId())
		courseInstance, _ := course.NewCourse(caddr, tf.Backend)

		var courseDigests [][32]byte
		for _, a := range c.Course.Assignments {
			// Publish digest of assignment credential
			digest := pb.Hash(a)
			courseDigests = append(courseDigests, digest)
			opts, _ := accounts.GetTxOpts(evaluators[0].Key, tf.Backend)
			_, err := courseInstance.RegisterCredential(opts, student.Address, digest)
			if err != nil {
				t.Fatalf("RegisterCredential expected no error, got: %v", err)
			}
			tf.Backend.Commit()
			proof := courseInstance.IssuedCredentials(nil, digest)
			assert.Equal(t, digest, proof.Digest)

			// Second evaluator confirms
			opts, _ = accounts.GetTxOpts(evaluators[1].Key, tf.Backend)
			_, err = courseInstance.RegisterCredential(opts, student.Address, digest)
			if err != nil {
				t.Fatalf("RegisterCredential expected no error, got: %v", err)
			}
			tf.Backend.Commit()

			opts, _ = accounts.GetTxOpts(student.Key, tf.Backend)
			_, err = courseInstance.ConfirmCredential(opts, digest)
			if err != nil {
				t.Fatalf("ConfirmCredential expected no error, got: %v", err)
			}
			tf.Backend.Commit()
		}

		// issue final course certificate
		digest := pb.Hash(c)
		courseDigests = append(courseDigests, digest)
		opts, _ := accounts.GetTxOpts(evaluators[0].Key, tf.Backend)
		_, err := courseInstance.RegisterCredential(opts, student.Address, digest)
		if err != nil {
			t.Fatalf("RegisterCredential expected no error, got: %v", err)
		}
		tf.Backend.Commit()
		proof := courseInstance.IssuedCredentials(nil, digest)
		assert.Equal(t, digest, proof.Digest)

		// Second evaluator confirms
		opts, _ = accounts.GetTxOpts(evaluators[1].Key, tf.Backend)
		_, err = courseInstance.RegisterCredential(opts, student.Address, digest)
		if err != nil {
			t.Fatalf("RegisterCredential expected no error, got: %v", err)
		}
		tf.Backend.Commit()

		opts, _ = accounts.GetTxOpts(student.Key, tf.Backend)
		_, err = courseInstance.ConfirmCredential(opts, digest)
		if err != nil {
			t.Fatalf("ConfirmCredential expected no error, got: %v", err)
		}
		tf.Backend.Commit()

		d, err := encode.EncodeByteArray(courseDigests)
		if err != nil {
			t.Error(err)
		}
		expectedDigests = append(expectedDigests, d)
	}
	// Finalize courses
	for _, c := range diploma.Courses {
		caddr := common.HexToAddress(c.Course.GetId())
		courseInstance, _ := course.NewCourse(caddr, tf.Backend)

		opts, _ := accounts.GetTxOpts(evaluators[0].Key, tf.Backend)
		_, err := courseInstance.AggregateCredentials(opts, student.Address)
		if err != nil {
			t.Fatalf("AggregateCredentials expected no error, got: %v", err)
		}
		tf.Backend.Commit()
	}

	diplomaCredential := pb.NewFakeDiplomaCredential(adms[0].Address.Hex(), diploma)
	digest := pb.Hash(diplomaCredential)
	expectedDigests = append(expectedDigests, digest)
	digestRoot, _ := encode.EncodeByteArray(expectedDigests)

	collectedDigests, err := tf.Faculty.CollectCredentials(&bind.CallOpts{From: adms[0].Address}, student.Address, coursesAddresses)
	if err != nil {
		t.Fatal(err)
	}
	root, _ := encode.EncodeByteArray(append(collectedDigests, digest))
	assert.Equal(t, digestRoot, root)

	opts, _ := accounts.GetTxOpts(adms[0].Key, tf.Backend)
	_, err = tf.Faculty.RegisterRootCredential(opts, student.Address, digest, digestRoot, coursesAddresses)
	if err != nil {
		t.Fatalf("RegisterRootCredential expected no error, got: %v", err)
	}
	tf.Backend.Commit()

	d := tf.Faculty.IssuedCredentials(nil, digest)
	assert.Equal(t, digest, d.Digest)

	// Second evaluator confirms
	opts, _ = accounts.GetTxOpts(adms[1].Key, tf.Backend)
	_, err = tf.Faculty.RegisterCredential(opts, student.Address, digest)
	if err != nil {
		t.Fatalf("RegisterCredential expected no error, got: %v", err)
	}
	tf.Backend.Commit()

	opts, _ = accounts.GetTxOpts(student.Key, tf.Backend)
	_, err = tf.Faculty.ConfirmCredential(opts, digest)
	if err != nil {
		t.Fatalf("ConfirmCredential expected no error, got: %v", err)
	}
	tf.Backend.Commit()

	if ok, _ := tf.Faculty.Certified(nil, digest); !ok {
		t.Fatalf("Digest %x should be certified", digest)
	}
}
