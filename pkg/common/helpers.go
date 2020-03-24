package common

import (
	"math/rand"
	"time"
)

var rkChars = []byte("abcdefghijklmnopqrstuvwxyz0123456789")
var r *rand.Rand

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// GenerateKey generates a random lowercase alphanumeric key.
//
// These keys approximate the Events API's routing keys for use in testing,
// but may be useful more generally.
func GenerateKey() string {
	rk := make([]byte, 32)
	for i, _ := range rk {
		rk[i] = randChar()
	}
	return string(rk)
}

func randChar() byte {
	return rkChars[r.Intn(len(rkChars))]
}
