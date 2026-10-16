// A first Go program for the Nintendo 3DS: it writes on the bottom screen,
// makes some garbage every frame so the collector has work, and goes back to
// the Homebrew Launcher when START is pressed.
//
//	tinygo build -target=nintendo3ds -o hello.3dsx ./src/examples/nintendo3ds-hello
package main

/*
#include <3ds.h>

static void startConsole(void) {
	gfxInitDefault();
	consoleInit(GFX_BOTTOM, NULL);
}

// nextFrame shows what was drawn, waits for the screen and reports whether
// START was pressed.
static int nextFrame(void) {
	gfxFlushBuffers();
	gfxSwapBuffers();
	gspWaitForVBlank();
	hidScanInput();
	return (hidKeysDown() & KEY_START) != 0;
}
*/
import "C"

import (
	"runtime"
	"time"
)

// The last blocks of garbage, kept so the collector must find them through the
// globals. A global, so the compiler cannot prove the blocks unused and drop
// them.
var ring [8][]byte

func main() {
	C.startConsole()
	println("Hello from Go on the 3DS")
	println("TinyGo + libctru")
	println()
	println("Press START to exit")

	var statistics runtime.MemStats
	start := time.Now()
	frames := 0
	for C.aptMainLoop() {
		if C.nextFrame() != 0 {
			break
		}
		block := make([]byte, 256*1024)
		block[0] = byte(frames)
		ring[frames%len(ring)] = block
		frames++
		if frames%60 != 0 {
			continue
		}
		if ring[(frames-1)%len(ring)][0] != byte(frames-1) {
			println("ERROR: the collector freed memory in use")
		}
		runtime.ReadMemStats(&statistics)
		print("\x1b[7;0H") // back to the 7th line of the console
		println("frames:     ", frames, "     ")
		println("seconds:    ", time.Since(start).Milliseconds()/1000, "     ")
		println("collections:", statistics.NumGC, "     ")
		println("heap:       ", statistics.HeapSys/1024/1024, "MB   ")
	}
	C.gfxExit()
}
