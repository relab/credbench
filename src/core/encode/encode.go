package encode

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
)

func EncodeByteArray(byteArray [][32]byte) ([32]byte, error) {
	var hash [32]byte

	bytes32ArrayType, err := abi.NewType("bytes32[]", "", nil)
	if err != nil {
		return hash, err
	}
	arguments := abi.Arguments{{Type: bytes32ArrayType}}
	bytes, err := arguments.Pack(byteArray)
	if err != nil {
		return hash, err
	}

	hash = crypto.Keccak256Hash(bytes)
	return hash, nil
}
