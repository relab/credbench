package client

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/r0qs/dvcp/src/course"
)

//FIXME: remove this function
// return the private key and the public key (address) based on the secp256k1  parameter of the elliptic curve of the primary key
func getKeys(hexkey string) (*ecdsa.PrivateKey, common.Address) {
	key, err := crypto.HexToECDSA(hexkey)
	if err != nil {
		log.Fatal(err)
	}
	address := crypto.PubkeyToAddress(key.PublicKey)
	return key, address
}

func Connect(host string) {
	client, err := ethclient.Dial(host)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	defer client.Close()

	// FIXME: remove this to a account manager class
	teacher, teacherAddress := getKeys("4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d")

	_, evaluatorAddress := getKeys("6cbed15c793ce57650b9877cf6fa156fbef513c4e6134f022a85b1ffdd59b2a1")

	// student, studentAddress := getKeys("6370fd033278c143179d81c5526140625662b8daa446c22ee2d73db3707e620c")

	// sender setup
	nonce, err := client.PendingNonceAt(context.Background(), teacherAddress)
	if err != nil {
		log.Fatal(err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	fmt.Println("gas price: ", gasPrice)
	if err != nil {
		log.Fatal(err)
	}

	auth := bind.NewKeyedTransactor(teacher)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(6721975)
	auth.GasPrice = gasPrice

	contractAddress, contractInstance, err := course.DeployCourse(auth, client, []common.Address{teacherAddress, evaluatorAddress}, big.NewInt(2))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(contractAddress.Hex())
	b, _ := contractInstance.Owners()
	for _, addr := range b {
		fmt.Println(addr.Hex())
	}
}
