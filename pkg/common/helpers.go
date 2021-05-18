package common

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"github.com/google/uuid"
)

var rkChars = []byte("abcdefghijklmnopqrstuvwxyz0123456789")
var r *rand.Rand

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func fileExists(f string) (bool, error) {
	_, err := os.Stat(f)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	} else {
		return true, nil
	}
}

func CreateAgentIdFile(agentIdFile string) bool {
	exists, err := fileExists(agentIdFile)
	if err != nil {
		fmt.Printf("Error checking if agent id file exists.\n")
		return false
	}

	if exists {
		return true
	}

	id := uuid.NewString()
	err = ioutil.WriteFile(agentIdFile, []byte(id), 0744)

	if err != nil {
		fmt.Printf("Error writing to agent id file.\n")
		return false
	}

	fmt.Printf("Successfully created agent id file.\n")
	return true
}

func GetAgentId(agentIdFile string) string {
	exists, err := fileExists(agentIdFile)
	if err != nil {
		fmt.Printf("Error checking if agent id file exists.\n")
		return "unavailable"
	}

	if !exists {
		created := CreateAgentIdFile(agentIdFile)
		if !created {
			fmt.Printf("Unable to get agent id.\n")
			return "unavailable"
		}
	}

	agentId, err := ioutil.ReadFile(agentIdFile)
	if err != nil {
		fmt.Printf("Unable to get agent id.")
		return "unavailable"
	}

	return string(agentId)
}

// GenerateKey generates a random lowercase alphanumeric key.
//
// These keys approximate the Events API's routing keys for use in testing,
// but may be useful more generally.
func GenerateKey() string {
	rk := make([]byte, 32)
	for i := range rk {
		rk[i] = randChar()
	}
	return string(rk)
}

func randChar() byte {
	return rkChars[r.Intn(len(rkChars))]
}
