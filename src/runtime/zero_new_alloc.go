//go:build gc.leaking && (baremetal || nintendoswitch || nintendo3ds)

package runtime

import (
	"unsafe"
)

//go:inline
func zero_new_alloc(ptr unsafe.Pointer, size uintptr) {
	memzero(ptr, size)
}
