//go:build none

// This file is included in the build for the Nintendo 3DS (despite the //go:build line above).

// Small wrappers around libctru functions that are static inline in its headers, so the Go
// runtime (which can only call real symbols) can reach them.

#include <3ds.h>

uint64_t tinygo_nintendo3ds_system_tick(void) {
    return svcGetSystemTick();
}
