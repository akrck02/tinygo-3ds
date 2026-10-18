package compileopts

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDevkitProFoldersComeFromTheEnvironment(t *testing.T) {
	devkitPro := t.TempDir()
	for _, version := range []string{"9.1.0", "16.1.0", "15.2.0"} {
		if err := os.MkdirAll(filepath.Join(devkitPro, "devkitARM", "lib", "gcc", "arm-none-eabi", version), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("DEVKITPRO", devkitPro)
	if got, want := expandTargetPath("-I{devkitpro}/libctru/include"), "-I"+devkitPro+"/libctru/include"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := expandTargetPath("{devkitarm-gcc}/armv6k/fpu/crti.o"), filepath.Join(devkitPro, "devkitARM", "lib", "gcc", "arm-none-eabi", "16.1.0")+"/armv6k/fpu/crti.o"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDevkitProsMSYSPathIsAWindowsFolder(t *testing.T) {
	cases := map[string]string{"": `C:\devkitPro`, "/opt/devkitpro": `C:\devkitPro`, `D:\tools\devkitPro`: `D:\tools\devkitPro`}
	for given, want := range cases {
		if got := windowsFolder(given); got != want {
			t.Errorf("%q gives %q, want %q", given, got, want)
		}
	}
}
