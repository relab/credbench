package course

import (
	"crypto/ecdsa"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/relab/ct-eth-dapp/src/accounts"
	"github.com/relab/ct-eth-dapp/src/backends"
	"github.com/relab/ct-eth-dapp/src/ctree/aggregator"
	"github.com/relab/ct-eth-dapp/src/ctree/notary"
	"github.com/relab/ct-eth-dapp/src/encode"

	pb "github.com/relab/ct-eth-dapp/src/schemes"
)

type TestCourse struct {
	Backend    *backends.TestBackend
	Evaluators []backends.Account
	Students   []backends.Account
	Course     *Course
}

func NewTestCourse(t *testing.T, evaluators backends.Accounts) *TestCourse {
	backend := backends.NewTestBackend()

	courseAddr, _, err := deployCourse(backend, evaluators[0].Key, evaluators.Addresses(), uint8(len(evaluators)))
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
	opts, _ := accounts.GetTxOpts(tc.Evaluators[0].Key, tc.Backend)
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

	ch := make(chan *CourseContractCredentialIssued)
	sub, _ := tc.Course.contract.WatchCredentialIssued(nil, ch, [][32]byte{digest}, []common.Address{to}, nil)
	defer func() {
		sub.Unsubscribe()
	}()

	opts, _ := accounts.GetTxOpts(tc.Evaluators[0].Key, tc.Backend)
	_, err := tc.Course.RegisterCredential(opts, to, digest, []common.Address{})
	if err != nil {
		t.Fatalf("RegisterCredential expected no error, got: %v", err)
	}
	tc.Backend.Commit()

	event := <-ch
	assert.Equal(t, digest, event.Digest)

	proof := tc.Course.GetCredentialProof(nil, digest)
	assert.Equal(t, digest, proof.Digest)

	return digest
}

func (tc *TestCourse) ConfirmTestCredential(t *testing.T, from *ecdsa.PrivateKey, digest [32]byte) {
	ch := make(chan *CourseContractCredentialSigned)
	sub, _ := tc.Course.contract.WatchCredentialSigned(nil, ch, nil, [][32]byte{digest})
	defer func() {
		sub.Unsubscribe()
	}()

	opts, _ := accounts.GetTxOpts(from, tc.Backend)
	_, err := tc.Course.ConfirmCredential(opts, digest)
	if err != nil {
		t.Fatalf("ConfirmCredential expected no error, got: %v", err)
	}
	tc.Backend.Commit()

	event := <-ch
	assert.Equal(t, accounts.GetAddress(from), event.Signer)
}

func deployLibs(opts *bind.TransactOpts, backend *backends.TestBackend) (map[string]string, error) {
	// opts := bind.NewKeyedTransactor(prvKey)
	libs := make(map[string]string)

	aggregatorAddr, _, _, err := aggregator.DeployCredentialSum(opts, backend)
	if err != nil {
		return libs, err
	}
	libs["CredentialSum"] = aggregatorAddr.Hex()

	notaryAddr, _, _, err := notary.DeployNotaryContract(opts, backend)
	if err != nil {
		return libs, err
	}
	libs["Notary"] = notaryAddr.Hex()
	return libs, nil
}

func deployCourse(backend *backends.TestBackend, prvKey *ecdsa.PrivateKey, evaluators []common.Address, quorum uint8) (common.Address, *Course, error) {
	opts := bind.NewKeyedTransactor(prvKey)

	libs, err := deployLibs(opts, backend)
	if err != nil {
		return common.Address{}, nil, err
	}

	courseAddr, _, c, err := DeployCourse(opts, backend, libs, evaluators, quorum)
	if err != nil {
		return common.Address{}, nil, err
	}
	backend.Commit()
	return courseAddr, c, nil
}

// Start tests

func TestNewCourse(t *testing.T) {
	evaluatorsAccount := backends.TestAccounts[:2]
	tc := NewTestCourse(t, evaluatorsAccount)
	defer tc.Backend.Close()

	// Calling Owners contract methods
	if ok, err := tc.Course.IsOwner(&bind.CallOpts{Pending: true}, tc.Evaluators[0].Address); !ok {
		t.Fatalf("IsOwner expected to be true but return: %t, %v", ok, err)
	}

	owners, err := tc.Course.GetOwners(&bind.CallOpts{Pending: true})
	if err != nil {
		t.Fatalf("OwnersList expected no errors but got: %v", err)
	}
	assert.ElementsMatch(t, evaluatorsAccount.Addresses(), owners)
}

func TestAddStudent(t *testing.T) {
	tc := NewTestCourse(t, backends.TestAccounts[:1])
	defer tc.Backend.Close()

	// Add a student
	studentAddress := backends.TestAccounts[2].Address
	opts, _ := accounts.GetTxOpts(tc.Evaluators[0].Key, tc.Backend)
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
	tc := NewTestCourse(t, backends.TestAccounts[:1])
	defer tc.Backend.Close()

	student := backends.TestAccounts[2]
	studentAddress := student.Address
	tc.AddStudents(t, backends.Accounts{student})

	// Remove a student
	opts, _ := accounts.GetTxOpts(tc.Evaluators[0].Key, tc.Backend)
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
	tc := NewTestCourse(t, backends.TestAccounts[:1])
	defer tc.Backend.Close()

	student := backends.TestAccounts[2]
	studentKey := student.Key
	studentAddress := student.Address
	tc.AddStudents(t, backends.Accounts{student})

	opts, _ := accounts.GetTxOpts(studentKey, tc.Backend)
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
	tc := NewTestCourse(t, backends.TestAccounts[:1])
	defer tc.Backend.Close()

	student := backends.TestAccounts[2]
	studentAddress := student.Address
	tc.AddStudents(t, backends.Accounts{student})

	digest := tc.RegisterTestCredential(t, studentAddress)

	proof := tc.Course.GetCredentialProof(nil, digest)

	assert.Equal(t, studentAddress, proof.Subject, "Subject address should be equal")
	assert.Equal(t, tc.Evaluators[0].Address, proof.Registrar, "Registrar address should be equal")
	assert.Equal(t, digest, proof.Digest, "Assignment digest should be equal")
	assert.False(t, proof.Approved)
}

func TestStudentSignCredential(t *testing.T) {
	tc := NewTestCourse(t, backends.TestAccounts[:1])
	defer tc.Backend.Close()

	student := backends.TestAccounts[2]
	studentKey := student.Key
	studentAddress := student.Address
	tc.AddStudents(t, backends.Accounts{student})

	digest := tc.RegisterTestCredential(t, studentAddress)

	proof := tc.Course.GetCredentialProof(nil, digest)
	assert.False(t, proof.Approved)

	tc.ConfirmTestCredential(t, studentKey, digest)

	proof = tc.Course.GetCredentialProof(nil, digest)
	assert.True(t, proof.Approved)
}

func TestIssuerAggregateCredential(t *testing.T) {
	tc := NewTestCourse(t, backends.TestAccounts[:1])
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

	opts, _ := accounts.GetTxOpts(tc.Evaluators[0].Key, tc.Backend)
	_, err := tc.Course.AggregateCredentials(opts, studentAddress, digests)
	if err != nil {
		t.Fatalf("AggregateCredentials expected no error, got: %v", err)
	}
	tc.Backend.Commit()

	aggregatedDigest, err := tc.Course.GetRoot(nil, studentAddress)
	if err != nil {
		t.Error(err)
	}

	expectedDigest, err := encode.EncodeByteArray(digests)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, expectedDigest, aggregatedDigest)
}

func VerifyCredentialTree(t *testing.T) {
	tc := NewTestCourse(t, backends.TestAccounts[:1])
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

	opts, _ := accounts.GetTxOpts(tc.Evaluators[0].Key, tc.Backend)
	_, err := tc.Course.AggregateCredentials(opts, studentAddress, digests)
	if err != nil {
		t.Fatalf("AggregateCredentials expected no error, got: %v", err)
	}
	tc.Backend.Commit()

	expectedDigest, err := encode.EncodeByteArray(digests)
	if err != nil {
		t.Error(err)
	}

	b, err := tc.Course.contract.VerifyCredentialRoot(nil, studentAddress, expectedDigest)
	if err != nil {
		t.Error(err)
	}
	assert.True(t, b)

	b, err = tc.Course.OnVerifyCredentialTree(nil, studentAddress)
	if err != nil {
		t.Error(err)
	}
	assert.True(t, b)
}
