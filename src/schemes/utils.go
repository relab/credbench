package schemes

import (
	"crypto/sha256"
	"fmt"
	proto "github.com/golang/protobuf/proto"
)

func Hash(pb proto.Message) [32]byte {
	data, _ := proto.Marshal(pb)
	return sha256.Sum256(data)
}

func HashString(s string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
}
