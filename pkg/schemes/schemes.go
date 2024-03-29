package schemes

import (
	"crypto/sha256"
	"io/ioutil"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	log "github.com/sirupsen/logrus"
)

func Hash(pb proto.Message) [32]byte {
	data, _ := proto.Marshal(pb)
	return sha256.Sum256(data)
}

func ParseJSON(path string, m proto.Message) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal("Error reading file:", err)
	}

	err = protojson.Unmarshal(data, m)
	if err != nil {
		log.Fatalf("unexpected error when unmarshaling json: %v", err)
	}
}
