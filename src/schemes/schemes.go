package schemes

import (
	"bytes"
	"crypto/sha256"
	"io/ioutil"
	"log"

	"github.com/golang/protobuf/jsonpb"
	proto "github.com/golang/protobuf/proto"
)

func Hash(pb proto.Message) [32]byte {
	data, _ := proto.Marshal(pb)
	return sha256.Sum256(data)
}

func ParseJSON(path string, m proto.Message) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalln("Error reading file:", err)
	}

	err = jsonpb.Unmarshal(bytes.NewReader(data), m)
	if err != nil {
		log.Fatalf("unexpected error when unmarshaling json: %v", err)
	}
}
