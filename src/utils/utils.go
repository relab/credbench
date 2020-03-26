package utils

import (
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
)

func CreateDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

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
