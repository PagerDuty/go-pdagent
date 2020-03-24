package common

import (
	"testing"
)

func TestGenerateRoutingKey(t *testing.T) {
	if len(GenerateKey()) != 32 {
		t.Error("Expected routing key to be exactly 32 characters.")
	}
}
