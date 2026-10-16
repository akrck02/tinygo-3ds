//go:build nintendo3ds

package runtime

// The Nintendo 3DS runs homebrew as .3dsx programs on top of devkitARM's newlib and libctru:
// their start code (3dsx_crt0) sets up the stack, the heap and the C library and then calls
// main, so this runtime only has to start Go, and it leans on newlib and libctru for the rest.

import "unsafe"

const zeroSizeAllocPtr uintptr = 16

// systemTickFrequency is how fast svcGetSystemTick counts: the ARM11 clock libctru calls
// SYSCLOCK_ARM11. It is the same on the original and the New 3DS.
const systemTickFrequency = 268111856

// The 3DS clock (osGetTime) counts milliseconds since 1 January 1900; Unix time starts 70 years
// later.
const millisecondsFrom1900To1970 = 2208988800 * 1000

// The heap is one block taken from newlib's malloc when the program starts. The first size
// that malloc can give is used, halving from the largest; the rest of the memory stays for C
// code, such as libctru and citro3d.
const (
	defaultHeapMegabytes = 64
	smallestHeapSize     = 1 * 1024 * 1024
)

// heapMegabytes, when set at build time, replaces the largest heap size: a smaller heap makes
// shorter collection pauses, since the collector sweeps all of it every cycle.
//
//	tinygo build -target=nintendo3ds -ldflags="-X runtime.heapMegabytes=32" ...
var heapMegabytes string

var heapStart, heapEnd uintptr

// stackTop is where the main thread's stack started when Go took over; the garbage collector
// scans from the current stack pointer up to it.
var stackTop uintptr

// Entry point, called by 3dsx_crt0 once the C library is ready.
//
//export main
func main() int32 {
	stackTop = getCurrentStackPointer()
	setupHeap()
	run()
	libc_exit(0)
	return 0
}

func setupHeap() {
	for size := largestHeapSize(); size >= smallestHeapSize; size /= 2 {
		block := libc_malloc(size)
		if block != nil {
			heapStart = uintptr(block)
			heapEnd = heapStart + size
			return
		}
	}
	runtimePanic("cannot reserve memory for the heap")
}

func largestHeapSize() uintptr {
	megabytes := uintptr(0)
	for index := 0; index < len(heapMegabytes); index++ {
		digit := heapMegabytes[index]
		if digit < '0' || digit > '9' {
			return defaultHeapMegabytes * 1024 * 1024
		}
		megabytes = megabytes*10 + uintptr(digit-'0')
	}
	if megabytes == 0 {
		megabytes = defaultHeapMegabytes
	}
	return megabytes * 1024 * 1024
}

// growHeap tries to grow the heap size. The heap is a single block, so it cannot grow.
func growHeap() bool {
	return false
}

func ticks() timeUnit {
	return timeUnit(systemTicksToNanoseconds(tinygo_nintendo3ds_system_tick()))
}

// systemTicksToNanoseconds splits the conversion so the product does not overflow 64 bits.
func systemTicksToNanoseconds(tick uint64) int64 {
	seconds := tick / systemTickFrequency
	remainder := tick % systemTickFrequency
	return int64(seconds*1000000000 + remainder*1000000000/systemTickFrequency)
}

// ticks are already nanoseconds.
func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks)
}

func nanosecondsToTicks(nanoseconds int64) timeUnit {
	return timeUnit(nanoseconds)
}

func sleepTicks(duration timeUnit) {
	svcSleepThread(int64(duration))
}

//go:linkname now time.now
func now() (sec int64, nsec int32, mono int64) {
	mono = nanotime()
	unixMilliseconds := int64(osGetTime()) - millisecondsFrom1900To1970
	sec = unixMilliseconds / 1000
	nsec = int32(unixMilliseconds%1000) * 1000000
	return
}

// Output is gathered by line and sent both to standard output (the libctru console, once a
// program has opened one) and to the debug log that emulators show.
var outputLine [120]byte
var outputLength = 0

func putchar(character byte) {
	if outputLength < len(outputLine) {
		outputLine[outputLength] = character
		outputLength++
	}
	if character == '\n' || outputLength == len(outputLine) {
		flushOutputLine()
	}
}

func flushOutputLine() {
	if outputLength == 0 {
		return
	}
	libc_write(1, &outputLine[0], outputLength)
	svcOutputDebugString(&outputLine[0], int32(outputLength))
	outputLength = 0
}

func buffered() int {
	return 0
}

func getchar() byte {
	return 0
}

func abort() {
	flushOutputLine()
	libc_exit(1)
	for {
	}
}

func exit(code int) {
	flushOutputLine()
	libc_exit(int32(code))
	for {
	}
}

// Go runs on a single thread and is never preempted, so nothing can move it to another
// processor while sync/atomic works.
//
//go:linkname procPin sync/atomic.runtime_procPin
func procPin() {
}

//go:linkname procUnpin sync/atomic.runtime_procUnpin
func procUnpin() {
}

//go:linkname syscall_Exit syscall.Exit
func syscall_Exit(code int) {
	exit(code)
}

// hardwareRand asks the system's ps service, the same source crypto/rand uses.
func hardwareRand() (n uint64, ok bool) {
	ok = tinygo_nintendo3ds_random(unsafe.Pointer(&n), 8) == 0
	return
}

func libc_errno_location() *int32 {
	return libc___errno()
}

//go:extern __start__
var imageStartSymbol [0]byte

//go:extern __exidx_end
var readOnlyEndSymbol [0]byte

//go:extern __end__
var imageEndSymbol [0]byte

// findGlobals scans the data and bss segments. devkitARM's linker script names no symbol at the
// start of .data, but places it at the first page after the read-only data ends.
func findGlobals(found func(start, end uintptr)) {
	const pageSize = 0x1000
	dataStart := (uintptr(unsafe.Pointer(&readOnlyEndSymbol)) + pageSize - 1) &^ (pageSize - 1)
	found(dataStart, uintptr(unsafe.Pointer(&imageEndSymbol)))
}

//export tinygo_nintendo3ds_system_tick
func tinygo_nintendo3ds_system_tick() uint64

//export tinygo_nintendo3ds_random
func tinygo_nintendo3ds_random(buffer unsafe.Pointer, size uint32) int32

//export svcSleepThread
func svcSleepThread(nanoseconds int64)

//export svcOutputDebugString
func svcOutputDebugString(text *byte, length int32) int32

//export osGetTime
func osGetTime() uint64

//export malloc
func libc_malloc(size uintptr) unsafe.Pointer

//export write
func libc_write(fd int32, buffer *byte, count int) int

//export exit
func libc_exit(code int32)

//export __errno
func libc___errno() *int32
