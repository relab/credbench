package accounts

import (
	"crypto/ecdsa"
	crand "crypto/rand"
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

func Has0xPrefix(input string) bool {
	return len(input) >= 2 && input[0] == '0' && (input[1] == 'x' || input[1] == 'X')
}

func HexToKey(hexkey string) *ecdsa.PrivateKey {
	if Has0xPrefix(hexkey) {
		hexkey = hexkey[2:]
	}
	key, err := crypto.HexToECDSA(hexkey)
	if err != nil {
		log.Panic(err)
	}
	return key
}

func NewKey() (*ecdsa.PrivateKey, common.Address) {
	privateKey, err := ecdsa.GenerateKey(crypto.S256(), crand.Reader)
	if err != nil {
		log.Panic(err)
	}
	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	return privateKey, address
}

func KeyToHex(privateKey *ecdsa.PrivateKey) string {
	keyBytes := crypto.FromECDSA(privateKey)
	return hexutil.Encode(keyBytes)
}

func GetKeys(hexkey string) (*ecdsa.PrivateKey, common.Address, error) {
	if Has0xPrefix(hexkey) {
		hexkey = hexkey[2:]
	}
	key, err := crypto.HexToECDSA(hexkey)
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("Error parsing the private key: %v", err)
	}
	address := crypto.PubkeyToAddress(key.PublicKey)
	return key, address, nil
}

// IsZeroAddress checks whether an address is 0
// Copied from:
// https://github.com/miguelmota/go-ethutil/blob/master/ethutil.go#L43
func IsZeroAddress(iaddress interface{}) bool {
	var address common.Address
	switch v := iaddress.(type) {
	case string:
		address = common.HexToAddress(v)
	case common.Address:
		address = v
	default:
		return false
	}

	addressBytes := address.Bytes()
	return reflect.DeepEqual(addressBytes, common.FromHex("0x0000000000000000000000000000000000000000"))
}
