//go:build nintendo3ds

package syscall

import (
	"internal/itoa"
)

// A Signal is a number describing a process signal.
// It implements the os.Signal interface.
type Signal int

const (
	_ Signal = iota
	SIGCHLD
	SIGINT
	SIGKILL
	SIGTRAP
	SIGQUIT
	SIGTERM
)

func (s Signal) Signal() {}

func (s Signal) String() string {
	if 0 <= s && int(s) < len(signals) {
		str := signals[s]
		if str != "" {
			return str
		}
	}
	return "signal " + itoa.Itoa(int(s))
}

var signals = [...]string{}

// File system

const (
	Stdin  = 0
	Stdout = 1
	Stderr = 2
)

// The values of devkitARM's newlib (sys/_default_fcntl.h), which differ from Linux.
const (
	O_RDONLY = 0
	O_WRONLY = 1
	O_RDWR   = 2

	O_APPEND = 0x0008
	O_CREAT  = 0x0200
	O_TRUNC  = 0x0400
	O_EXCL   = 0x0800
	O_SYNC   = 0x2000

	O_CLOEXEC = 0
)

// newlib keeps errno per thread and hands out its address through __errno.
//
//export __errno
func libc_errno_location() *int32

func getErrno() error {
	return Errno(uintptr(*libc_errno_location()))
}

func Pipe2(p []int, flags int) (err error) {
	return ENOSYS // TODO
}

func Getpagesize() int {
	return 4096 // TODO
}

type RawSockaddrInet4 struct {
	// stub
}

type RawSockaddrInet6 struct {
	// stub
}

func Chmod(path string, mode uint32) (err error) {
	data := cstring(path)
	fail := int(libc_chmod(&data[0], mode))
	if fail < 0 {
		err = getErrno()
	}
	return
}

// int open(const char *pathname, int flags, mode_t mode);
//
//export open
func libc_open(pathname *byte, flags int32, mode uint32) int32
