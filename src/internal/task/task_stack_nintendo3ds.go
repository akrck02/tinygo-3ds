//go:build scheduler.tasks && nintendo3ds

package task

// The Nintendo 3DS uses the hard-float ABI, where d8-d15 are callee-saved
// like r4-r11, so a task switch saves them too (see task_stack_nintendo3ds.S).
// Otherwise this is task_stack_arm.go.

import "unsafe"

var systemStack uintptr

// calleeSavedRegs is the list of registers that must be saved and restored when
// switching between tasks, in the order task_stack_nintendo3ds.S pushes them:
// vpush {d8-d15} below push {r3-r11, lr}. r3 is not callee-saved; it is only
// there to keep the block, and so the stack pointer, a multiple of 8 bytes.
type calleeSavedRegs struct {
	d8, d9, d10, d11, d12, d13, d14, d15 uint64

	r3  uintptr
	r4  uintptr
	r5  uintptr
	r6  uintptr
	r7  uintptr
	r8  uintptr
	r9  uintptr
	r10 uintptr
	r11 uintptr

	pc uintptr
}

// archInit runs architecture-specific setup for the goroutine startup.
func (s *state) archInit(r *calleeSavedRegs, fn uintptr, args unsafe.Pointer) {
	// Store the initial sp for the startTask function (implemented in assembly).
	s.sp = uintptr(unsafe.Pointer(r))

	// Start the function at tinygo_startTask, which calls the function in r4
	// with the argument in r5 and then exits the goroutine.
	r.pc = uintptr(unsafe.Pointer(&startTask))
	r.r4 = fn
	r.r5 = uintptr(args)
}

func (s *state) resume() {
	swapTask(s.sp, &systemStack)
}

func (s *state) pause() {
	newStack := systemStack
	systemStack = 0
	swapTask(newStack, &s.sp)
}

// SystemStack returns the system stack pointer when called from a task stack.
// When called from the system stack, it returns 0.
func SystemStack() uintptr {
	return systemStack
}
