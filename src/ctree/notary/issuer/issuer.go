package issuer

//go:generate abigen --abi ../../../ethereum/build/abi/Issuer.abi --bin ../../../ethereum/build/bin/Issuer.bin --pkg issuer --type IssuerContract --out ./issuer_contract.go

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/relab/ct-eth-dapp/src/ctree/owners"
)

// Issuer is a Go wrapper around an issuer contract.
type Issuer struct {
	*owners.Owners
	address  common.Address
	contract *IssuerContract
}

// NewIssuer creates a struct exposing convenient operations to
// interact with the Issuer contract.
func NewIssuer(contractAddr common.Address, backend bind.ContractBackend) (*Issuer, error) {
	i, err := NewIssuerContract(contractAddr, backend)
	if err != nil {
		return nil, err
	}
	o, err := owners.NewOwners(contractAddr, backend)
	if err != nil {
		return nil, err
	}
	return &Issuer{o, contractAddr, i}, nil
}

// Address returns the contract address of the issuer.
func (i *Issuer) Address() common.Address {
	return i.address
}

// GetCredentialProof maps document digest to issued credential proof
func (i *Issuer) GetCredentialProof(opts *bind.CallOpts, digest [32]byte) *NotaryCredentialProof {
	proof, _ := i.contract.GetCredentialProof(opts, digest)
	var cp NotaryCredentialProof = proof
	return &cp
}

// GetRevokedProof maps document digest to revoked proof
func (i *Issuer) GetRevokedProof(opts *bind.CallOpts, digest [32]byte) *NotaryRevocationProof {
	proof, _ := i.contract.GetRevokedProof(opts, digest)
	var rp NotaryRevocationProof = proof
	return &rp
}

// IsSigned returns whether an owner already signed a digest
func (i *Issuer) IsSigned(opts *bind.CallOpts, digest [32]byte, owner common.Address) (bool, error) {
	return i.contract.IsSigned(opts, digest, owner)
}

// IsQuorumSigned verify if a credential proof was signed by a quorum
func (i *Issuer) IsQuorumSigned(opts *bind.CallOpts, digest [32]byte) (bool, error) {
	return i.contract.IsQuorumSigned(opts, digest)
}

// GetDigests returns the list of the issued credentials' digests of a subject
func (i *Issuer) GetDigests(opts *bind.CallOpts, subject common.Address) ([][32]byte, error) {
	return i.contract.GetDigests(opts, subject)
}

// GetWitnesses returns the witnesses of a proof
func (i *Issuer) GetWitnesses(opts *bind.CallOpts, digest [32]byte) ([]common.Address, error) {
	return i.contract.GetWitnesses(opts, digest)
}

// GetEvidenceRoot returns the root of the evidences of an issued credential proof
func (i *Issuer) GetEvidenceRoot(opts *bind.CallOpts, digest [32]byte) ([32]byte, error) {
	return i.contract.GetEvidenceRoot(opts, digest)
}

// GetRevoked returns a list of revoked credentials
func (i *Issuer) GetRevoked(opts *bind.CallOpts, subject common.Address) ([][32]byte, error) {
	return i.contract.GetRevoked(opts, subject)
}

// IsRevoked verifies if a credential proof was revoked
func (i *Issuer) IsRevoked(opts *bind.CallOpts, digest [32]byte) (bool, error) {
	return i.contract.IsRevoked(opts, digest)
}

// OnVerifyCredential checks whether the credential is valid (on-chain).
func (i *Issuer) OnVerifyCredential(opts *bind.CallOpts, subject common.Address, digest [32]byte) (bool, error) {
	return i.contract.VerifyCredential(opts, subject, digest)
}

func (i *Issuer) OffVerifyCredential(opts *bind.CallOpts, subject common.Address, digest [32]byte) (bool, error) {
	cp, err := i.contract.GetCredentialProof(opts, digest)
	if err != nil {
		return false, err
	}
	if cp.InsertedBlock == big.NewInt(0) {
		return false, fmt.Errorf("Credential %x not found", digest)
	}
	if cp.Subject != subject {
		return false, fmt.Errorf("Wrong subject %s", subject.Hex())
	}
	if !cp.Approved {
		return false, fmt.Errorf("Credential not approved")
	}
	signed, err := i.contract.IsQuorumSigned(opts, digest)
	if err != nil {
		return false, err
	}
	if !signed {
		q, err := i.contract.Quorum(opts)
		if err != nil {
			return false, err
		}
		return false, fmt.Errorf("Credential not signed by quorum. Require %d but has %d", q, cp.Signed)
	}
	revoked, err := i.contract.IsRevoked(opts, digest)
	if revoked {
		rp, err := i.contract.GetRevokedProof(opts, digest)
		if err != nil {
			return false, err
		}
		return false, fmt.Errorf("Credential revoked at %s", rp.RevokedBlock.String())
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// OnVerifyIssuedCredentials checks whether all credentials of a given subject are valid (on-chain).
func (i *Issuer) OnVerifyIssuedCredentials(opts *bind.CallOpts, subject common.Address) (bool, error) {
	return i.contract.VerifyIssuedCredentials(opts, subject)
}

func (i *Issuer) OffVerifyIssuedCredentials(opts *bind.CallOpts, subject common.Address) (bool, error) {
	digests, err := i.contract.GetDigests(opts, subject)
	if err != nil {
		return false, err
	}
	for _, d := range digests {
		if ok, err := i.OffVerifyCredential(opts, subject, d); !ok || err != nil {
			return false, err
		}
	}
	return true, nil
}
