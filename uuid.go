package nacre

import (
	"crypto/rand"
	"fmt"
)

// NewUUID returns a new universally unique identifier.
func NewUUID() string {
	buf := make([]byte, 16)
	_, err := rand.Read(buf)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf(
		"%x-%x-%x-%x",
		buf[0:4],
		buf[4:8],
		buf[8:12],
		buf[12:16],
	)
}
