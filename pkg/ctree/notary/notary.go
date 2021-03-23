package notary

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// NotaryCredentialProof
type NotaryCredentialProof struct {
	Signed         *big.Int
	InsertedBlock  *big.Int
	BlockTimestamp *big.Int
	Nonce          *big.Int
	Digest         [32]byte
	Approved       bool
	Registrar      common.Address
	Subject        common.Address
	Witnesses      []common.Address
	EvidenceRoot   [32]byte
}

// NotaryRevocationProof
type NotaryRevocationProof struct {
	Registrar    common.Address
	Subject      common.Address
	RevokedBlock *big.Int
	Reason       [32]byte
}
