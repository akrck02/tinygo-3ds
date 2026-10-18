package compileopts

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/tinygo-org/tinygo/goenv"
)

// DevkitProFolder is devkitPro's folder: $DEVKITPRO, or where devkitPro installs itself.
func DevkitProFolder() string {
	if folder := os.Getenv("DEVKITPRO"); folder != "" {
		return folder
	}
	if runtime.GOOS == "windows" {
		return `C:\devkitPro`
	}
	return "/opt/devkitpro"
}

// devkitARMGCCFolder is the folder of devkitARM's newest GCC, which holds libgcc and the crt
// objects: its version changes with devkitARM's updates.
func devkitARMGCCFolder(devkitPro string) string {
	versions, _ := filepath.Glob(filepath.Join(devkitPro, "devkitARM", "lib", "gcc", "arm-none-eabi", "*"))
	if len(versions) == 0 {
		return filepath.Join(devkitPro, "devkitARM", "lib", "gcc", "arm-none-eabi", "unknown")
	}
	sort.Slice(versions, func(first, second int) bool { return newerVersion(versions[second], versions[first]) })
	return versions[len(versions)-1]
}

// newerVersion tells whether a folder named by a dotted version is newer than another.
func newerVersion(folder, other string) bool {
	parts, others := strings.Split(filepath.Base(folder), "."), strings.Split(filepath.Base(other), ".")
	for index := 0; index < len(parts) && index < len(others); index++ {
		if len(parts[index]) != len(others[index]) {
			return len(parts[index]) > len(others[index])
		}
		if parts[index] != others[index] {
			return parts[index] > others[index]
		}
	}
	return len(parts) > len(others)
}

// expandTargetPath fills the folders a target's flags name: {root} (TinyGo's), {devkitpro}
// (devkitPro's, from $DEVKITPRO) and {devkitarm-gcc} (devkitARM's GCC).
func expandTargetPath(flag string) string {
	flag = strings.ReplaceAll(flag, "{root}", goenv.Get("TINYGOROOT"))
	if !strings.Contains(flag, "{devkit") {
		return flag
	}
	devkitPro := DevkitProFolder()
	flag = strings.ReplaceAll(flag, "{devkitarm-gcc}", devkitARMGCCFolder(devkitPro))
	return strings.ReplaceAll(flag, "{devkitpro}", devkitPro)
}
