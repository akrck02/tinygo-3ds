//go:build tinygo

// task_stack_arm.S, saving the floating point registers d8-d15 as well: the
// Nintendo 3DS uses the hard-float ABI, where they are callee-saved.

// Only generate .debug_frame, don't generate .eh_frame.
.cfi_sections .debug_frame

.section .text.tinygo_startTask
.global  tinygo_startTask
.type    tinygo_startTask, %function
tinygo_startTask:
    .cfi_startproc
    // Run on the new stack with the callee-saved registers loaded: r4 holds the
    // function to start and r5 its only argument.
    .cfi_undefined lr
    mov   r0, r5
    blx   r4

    // After return, exit this goroutine. This is a tail call.
    bl    tinygo_task_exit
    .cfi_endproc
.size tinygo_startTask, .-tinygo_startTask

.section .text.tinygo_swapTask
.global tinygo_swapTask
.type tinygo_swapTask, %function
tinygo_swapTask:
    // r0 = newStack uintptr
    // r1 = oldStack *uintptr

    // Save the callee-saved registers (r3 only keeps 8-byte alignment).
    push  {r3-r11, lr}
    vpush {d8-d15}

    // Save the current stack pointer in oldStack.
    str sp, [r1]

    // Switch to the new stack pointer.
    mov sp, r0

    // Load state from the new task and branch to where it stopped.
    vpop  {d8-d15}
    pop   {r3-r11, pc}
.size tinygo_swapTask, .-tinygo_swapTask
