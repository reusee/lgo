package lgo

import (
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math/rand"
)

var (
	pt = fmt.Printf
)

func init() {
	var seed int64
	binary.Read(crand.Reader, binary.LittleEndian, &seed)
	rand.Seed(seed)
}
