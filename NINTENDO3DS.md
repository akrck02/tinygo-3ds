# Nintendo 3DS (homebrew)

Work in progress: a `nintendo3ds` target that builds Go programs as `.3dsx`
files for the Homebrew Launcher, on top of devkitPro's devkitARM, newlib and
libctru.

## What you need

- TinyGo built from this branch (`go install -tags=llvm20 .` against a system
  LLVM works).
- devkitPro with the 3DS packages (devkitARM, libctru, 3dstools) in
  `/opt/devkitpro`. The target's paths assume that location and devkitARM's
  GCC 16.1.0.

## Building

```bash
tinygo build -target=nintendo3ds -o hello.3dsx ./src/examples/nintendo3ds-hello
```

Copy the `.3dsx` to `/3ds/` on the SD card and open it from the Homebrew
Launcher, or open it in the Azahar emulator.

## How it works

- devkitARM's start code (`3dsx_crt0`) prepares the stack, the C library and
  libctru, then calls `main` in `src/runtime/runtime_nintendo3ds.go`, which
  starts Go.
- C code comes from devkitARM (newlib, libctru), not from TinyGo's own libc.
  Programs reach libctru through cgo; `#include <3ds.h>` works.
- The heap is one block taken from `malloc` at start: the largest the console
  gives, less 8 MB left for C code (about 86 MB on a New 3DS XL). Goroutine
  stacks are 128 KB: deep code (encoding/json, a large interface) overflowed
  16 KB, and an overflow writes over whatever is next to the stack.
- Output (`println`) goes to standard output, which is the libctru console once
  a program opens one, and to the debug log emulators show.
- `ld.lld` links the program with `targets/nintendo3ds.ld`, devkitARM's
  `3dsx.ld` with a TLS segment added, which `ld.lld` needs. The `.3dsx` output
  format runs devkitPro's `3dsxtool`.
- The CPU is the ARM11 MPCore (ARMv6K, VFPv2, hard-float ABI), the same as the
  one devkitARM builds for.
- Goroutines use the `tasks` scheduler, which is cooperative: a program's loop
  that waits in C (for the screen) must call `runtime.Gosched` now and then, or
  its other goroutines never run.
- `min` and `max` of floating point numbers are compares and selects on 32-bit
  ARM: LLVM cannot select its `llvm.minimum`/`llvm.maximum` there.
- Goroutines use the `tasks` scheduler. The task switch
  (`src/internal/task/task_stack_nintendo3ds.S`) also saves d8-d15, which the
  hard-float ABI keeps across calls.
- Random numbers (`crypto/rand`, the runtime's seed, `hash/maphash`) come from
  the system's ps service (`PS_GenerateRandomBytes`).

## Tests

`tinygo test -target=nintendo3ds -c -o tests.3dsx ./package` builds a test
program; it prints its results to the debug log, which the Azahar emulator
shows. To pass flags, set them at build time:
`-ldflags="-X runtime.osArgs=-test.v"`.

## Memory

The collector scans stacks conservatively and does not move what it keeps. On
the console both matter: a stale word on a stack can keep a whole decoded JSON
tree alive, and a program that makes and drops many small objects leaves the
heap in pieces, so a large allocation can fail with plenty free in total. Keep
large buffers and reuse them, and keep the data a program reads small.

## Not yet

- Files: `os` has no file system yet (like on the Switch), although newlib can
  reach the SD card. `os.ReadFile`, `os.WriteFile` and `t.TempDir` fail.
- Only the main thread; no use of the second core.
- Errno values are Linux's, not newlib's, for anything above the common ones.
- Collection pauses are long (hundreds of milliseconds with about 10 MB alive).
  The target uses the `precise` collector, which skips blocks without pointers
  (such as mesh data), and a smaller heap shortens the sweep:
  `-ldflags="-X runtime.heapMegabytes=32"`.
