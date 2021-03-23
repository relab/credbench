package faculty

import (
	"context"
	"crypto/ecdsa"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/relab/ct-eth-dapp/pkg/backends"
	"github.com/relab/ct-eth-dapp/pkg/course"
	"github.com/relab/ct-eth-dapp/pkg/ctree/node"
	"github.com/relab/ct-eth-dapp/pkg/encode"

	pb "github.com/relab/ct-eth-dapp/pkg/schemes"
)

type TestFaculty struct {
	Backend    *backends.TestBackend
	Adms       []backends.Account
	Evaluators []backends.Account
	Students   []backends.Account
	Faculty    *Faculty
	Diplomas   map[string]pb.DiplomaCredential // maps student address to certificates
}

func NewTestFaculty(t *testing.T, adms backends.Accounts, quorum uint8) *TestFaculty {
	backend := backends.NewTestBackend()

	facultyAddr, _, err := deployFaculty(backend, adms[0].Key, adms.Addresses(), uint8(len(adms)))
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

func deployFaculty(backend *backends.TestBackend, prvKey *ecdsa.PrivateKey, adms []common.Address, quorum uint8) (common.Address, *Faculty, error) {
	opts := bind.NewKeyedTransactor(prvKey)

	libs, err := backend.DeployLibs(opts)
	if err != nil {
		return common.Address{}, nil, err
	}
	facultyAddr, _, faculty, err := DeployFaculty(opts, backend, libs, adms, quorum)
	if err != nil {
		return common.Address{}, nil, err
	}
	backend.Commit()
	return facultyAddr, faculty, nil
}

func TestNewFaculty(t *testing.T) {
	admsAccount := backends.TestAccounts[:2]
	tf := NewTestFaculty(t, admsAccount, 2)
	defer tf.Backend.Close()

	if ok, err := tf.Faculty.IsOwner(&bind.CallOpts{Pending: true}, tf.Adms[0].Address); !ok {
		t.Fatalf("IsOwner expected to be true but return: %t, %v", ok, err)
	}

	adms, err := tf.Faculty.GetOwners(&bind.CallOpts{Pending: true})
	if err != nil {
		t.Fatalf("GetOwners expected no errors but got: %v", err)
	}
	assert.ElementsMatch(t, admsAccount.Addresses(), adms)
}

func TestFacultyAddCourse(t *testing.T) {
	adms := backends.TestAccounts[:2]
	evaluators := backends.TestAccounts[2:4]

	tf := NewTestFaculty(t, adms, uint8(len(adms)))
	defer tf.Backend.Close()

	opts := bind.NewKeyedTransactor(adms[0].Key)
	courseAddr, _, _, err := course.DeployCourse(opts, tf.Backend, tf.Backend.GetLibs(), evaluators.Addresses(), uint8(len(evaluators)))
	if err != nil {
		t.Fatalf("Failed to deploy course: %v", err)
	}
	tf.Backend.Commit()

	_, err = tf.Faculty.AddNode(opts, courseAddr)
	if err != nil {
		t.Fatalf("Failed to add child node: %v", err)
	}
	tf.Backend.Commit()

	block, _ := tf.Backend.BlockByNumber(context.TODO(), nil)
	query := ethereum.FilterQuery{
		ToBlock: block.Number(),
		Addresses: []common.Address{
			tf.Faculty.Address(),
		},
	}

	logs, err := tf.Backend.FilterLogs(context.TODO(), query)
	if err != nil {
		t.Fatal(err)
	}

	evNodeAdded, err := tf.Faculty.contract.ParseNodeAdded(logs[0])
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, evNodeAdded.CreatedBy, adms[0].Address)
	assert.Equal(t, evNodeAdded.NodeAddress, courseAddr)
	assert.Equal(t, evNodeAdded.Role, node.LeafRole)
}

func TestCreateDiploma(t *testing.T) {
	adms := backends.TestAccounts[:2]
	evaluators := backends.TestAccounts[2:4]
	student := backends.TestAccounts[4]

	tf := NewTestFaculty(t, adms, uint8(len(adms)))
	defer tf.Backend.Close()

	var coursesAddresses []common.Address
	for i := 0; i < 4; i++ {
		// adm creates course
		opts := bind.NewKeyedTransactor(adms[0].Key)
		courseAddr, _, _, err := course.DeployCourse(opts, tf.Backend, tf.Backend.GetLibs(), evaluators.Addresses(), uint8(len(evaluators)))
		if err != nil {
			t.Fatalf("Failed to deploy course: %v", err)
		}
		tf.Backend.Commit()
		coursesAddresses = append(coursesAddresses, courseAddr)

		_, err = tf.Faculty.AddNode(opts, courseAddr)
		if err != nil {
			t.Fatalf("Failed to add child node: %v", err)
		}
		tf.Backend.Commit()

		courseInstance, err := course.NewCourse(courseAddr, tf.Backend)
		if err != nil {
			t.Fatalf("Failed to get new course instance: %v", err)
		}
		if ok, _ := courseInstance.IsOwner(nil, evaluators[0].Address); !ok {
			t.Fatalf("Evaluator %v is expected to be owner of the course contract", evaluators[0].Address.Hex())
		}

		// Adding a student
		opts = bind.NewKeyedTransactor(evaluators[0].Key)
		_, err = courseInstance.AddStudent(opts, student.Address)
		if err != nil {
			t.Fatalf("Failed to add student to course %s: %v", courseInstance.Address().Hex(), err)
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

	rootPerCourse := make(map[common.Address][32]byte)
	courseDigests := make(map[common.Address][][32]byte)
	for _, c := range diploma.Courses {
		caddr := common.HexToAddress(c.Course.GetId())
		courseInstance, _ := course.NewCourse(caddr, tf.Backend)

		for _, a := range c.Course.Assignments {
			// Publish digest of assignment credential
			digest := pb.Hash(a)
			courseDigests[caddr] = append(courseDigests[caddr], digest)
			opts := bind.NewKeyedTransactor(evaluators[0].Key)
			_, err := courseInstance.RegisterCredential(opts, student.Address, digest, []common.Address{})
			if err != nil {
				t.Fatalf("RegisterCredential expected no error, got: %v", err)
			}
			tf.Backend.Commit()
			proof, err := courseInstance.GetCredentialProof(nil, digest)
			if err != nil {
				t.Fatalf("GetCredentialProof expected no error, got: %v", err)
			}
			assert.Equal(t, digest, proof.Digest)

			// Second evaluator confirms
			opts = bind.NewKeyedTransactor(evaluators[1].Key)
			_, err = courseInstance.RegisterCredential(opts, student.Address, digest, []common.Address{})
			if err != nil {
				t.Fatalf("RegisterCredential expected no error, got: %v", err)
			}
			tf.Backend.Commit()

			opts = bind.NewKeyedTransactor(student.Key)
			_, err = courseInstance.ApproveCredential(opts, digest)
			if err != nil {
				t.Fatalf("ApproveCredential expected no error, got: %v", err)
			}
			tf.Backend.Commit()
		}

		// issue final course certificate
		digest := pb.Hash(c)
		courseDigests[caddr] = append(courseDigests[caddr], digest)
		opts := bind.NewKeyedTransactor(evaluators[0].Key)
		_, err := courseInstance.RegisterCredential(opts, student.Address, digest, []common.Address{})
		if err != nil {
			t.Fatalf("RegisterCredential expected no error, got: %v", err)
		}
		tf.Backend.Commit()
		proof, err := courseInstance.GetCredentialProof(nil, digest)
		if err != nil {
			t.Fatalf("GetCredentialProof expected no error, got: %v", err)
		}
		assert.Equal(t, digest, proof.Digest)

		// Second evaluator also signs the credential
		opts = bind.NewKeyedTransactor(evaluators[1].Key)
		_, err = courseInstance.RegisterCredential(opts, student.Address, digest, []common.Address{})
		if err != nil {
			t.Fatalf("RegisterCredential expected no error, got: %v", err)
		}
		tf.Backend.Commit()

		opts = bind.NewKeyedTransactor(student.Key)
		_, err = courseInstance.ApproveCredential(opts, digest)
		if err != nil {
			t.Fatalf("ApproveCredential expected no error, got: %v", err)
		}
		tf.Backend.Commit()

		d, err := encode.EncodeByteArray(courseDigests[caddr])
		if err != nil {
			t.Error(err)
		}
		rootPerCourse[caddr] = d
	}
	// Finalize courses
	for _, c := range diploma.Courses {
		caddr := common.HexToAddress(c.Course.GetId())
		courseInstance, _ := course.NewCourse(caddr, tf.Backend)

		opts := bind.NewKeyedTransactor(evaluators[0].Key)
		_, err := courseInstance.AggregateCredentials(opts, student.Address, courseDigests[caddr])
		if err != nil {
			t.Fatalf("Failed to aggregate course credentials: %v", err)
		}
		tf.Backend.Commit()

		root, err := courseInstance.GetRoot(nil, student.Address)
		if err != nil {
			t.Fatalf("Root not found for course %s: %v", caddr.Hex(), err)
		}

		assert.Equal(t, rootPerCourse[caddr], root)
	}

	diplomaCredential := pb.NewFakeDiplomaCredential(adms[0].Address.Hex(), diploma)
	digest := pb.Hash(diplomaCredential)

	opts := bind.NewKeyedTransactor(adms[0].Key)
	_, err := tf.Faculty.RegisterCredential(opts, student.Address, digest, coursesAddresses)
	if err != nil {
		t.Fatalf("RegisterRootCredential expected no error, got: %v", err)
	}
	tf.Backend.Commit()

	// Should preserve the order of the courses used as witnesses when aggregating
	var roots [][32]byte
	for _, r := range coursesAddresses {
		roots = append(roots, rootPerCourse[r])
	}

	digestRoot, _ := encode.EncodeByteArray(roots)
	root, _ := tf.Faculty.GetEvidenceRoot(nil, digest)
	assert.Equal(t, digestRoot, root)

	d := tf.Faculty.GetCredentialProof(nil, digest)
	assert.Equal(t, digest, d.Digest)

	// Second administration staff confirm the diploma credentail
	opts = bind.NewKeyedTransactor(adms[1].Key)
	_, err = tf.Faculty.RegisterCredential(opts, student.Address, digest, coursesAddresses)
	if err != nil {
		t.Fatalf("Failed to register diploma credential: %v", err)
	}
	tf.Backend.Commit()

	opts = bind.NewKeyedTransactor(student.Key)
	_, err = tf.Faculty.ApproveCredential(opts, digest)
	if err != nil {
		t.Fatalf("failed to confirm issued credential: %v", err)
	}
	tf.Backend.Commit()

	if ok, _ := tf.Faculty.IsQuorumSigned(nil, digest); !ok {
		t.Fatalf("Digest %x should be signed", digest)
	}
}
