package builder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// makeNintendo3DSProgram turns an ELF file into a .3dsx program, the format the
// Homebrew Launcher runs, with 3dsxtool from devkitPro. The tool is looked up
// in $DEVKITPRO/tools/bin (or /opt/devkitpro/tools/bin) and then in $PATH.
func makeNintendo3DSProgram(elfPath, programPath string) error {
	tool, err := find3dsxtool()
	if err != nil {
		return err
	}
	command := exec.Command(tool, elfPath, programPath)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("3dsxtool failed: %w", err)
	}
	return nil
}

func find3dsxtool() (string, error) {
	devkitPro := os.Getenv("DEVKITPRO")
	if devkitPro == "" {
		devkitPro = "/opt/devkitpro"
	}
	candidate := filepath.Join(devkitPro, "tools", "bin", "3dsxtool")
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}
	if path, err := exec.LookPath("3dsxtool"); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("3dsxtool not found: install devkitPro's 3dstools and set DEVKITPRO")
}
