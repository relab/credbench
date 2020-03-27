package utils

import (
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
)

func HashString(s string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
}

func CreateDir(path string) (err error) {
	if _, err = os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateFile(filename string) (err error) {
	if _, err = os.Stat(filename); os.IsNotExist(err) {
		f, err := os.Create(filename)
		if err != nil {
			return err
		}
		f.Close()
	}
	return err
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
