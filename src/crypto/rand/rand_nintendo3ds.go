//go:build nintendo3ds

// This implementation of crypto/rand uses the Nintendo 3DS's ps service
// (PS_GenerateRandomBytes in libctru), through the runtime's C helper.

package rand

import (
	"errors"
	"unsafe"
)

func init() {
	Reader = &reader{}
}

type reader struct {
}

var errNoRandomness = errors.New("crypto/rand: the ps service gave no random bytes")

func (r *reader) Read(b []byte) (n int, err error) {
	if len(b) == 0 {
		return
	}
	if tinygo_nintendo3ds_random(unsafe.Pointer(&b[0]), uint32(len(b))) != 0 {
		return 0, errNoRandomness
	}
	return len(b), nil
}

//export tinygo_nintendo3ds_random
func tinygo_nintendo3ds_random(buffer unsafe.Pointer, size uint32) int32
