//go:build none

// This file is included in the build for the Nintendo 3DS (despite the //go:build line above).

// Small wrappers around libctru functions that are static inline in its headers, so the Go
// runtime (which can only call real symbols) can reach them.

#include <3ds.h>

uint64_t tinygo_nintendo3ds_system_tick(void) {
    return svcGetSystemTick();
}

// tinygo_nintendo3ds_random fills buffer with random bytes from the system's ps service, which
// it opens on first use. It returns 0 on success.
int tinygo_nintendo3ds_random(void *buffer, uint32_t size) {
    static bool opened = false;
    if (!opened) {
        if (R_FAILED(psInit())) {
            return -1;
        }
        opened = true;
    }
    return R_FAILED(PS_GenerateRandomBytes(buffer, size)) ? -1 : 0;
}
