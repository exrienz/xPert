package orchestrator

import (
	"crypto/rand"
	"encoding/hex"
)

func newID() string {
	buffer := make([]byte, 12)
	_, _ = rand.Read(buffer)
	return hex.EncodeToString(buffer)
}
