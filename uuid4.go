package routers

import (
	"fmt"
	"math/rand"
	"time"
)

// ---- UUID GENERATION ----

// UUID ... Wrapper for string type in uuid4 generator
type UUID string

// uuid4 ... Generate a random UUID v4 (NOT RFC 4122 COMPLIANT)
func uuid4() (s UUID, err error) {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 16)
	_, err = rand.Read(b)
	if err != nil {
		return
	}
	s = UUID(fmt.Sprintf("%X-%04X-%04X-%04X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]))
	return
}
