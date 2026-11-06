//go:build windows

package shell

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Default returns the shell commands run in. Models write bash far more
// reliably than PowerShell, and the agent's command tool is called bash,
// so Git Bash wins when installed. Next comes PowerShell 7, which supports
// && and ||, then Windows PowerShell, which every Windows has. $SHELL is
// ignored: when set at all it usually names an MSYS path that cannot be
// started directly.
func Default() string {
	if bash := gitBash(); bash != "" {
		return bash
	}
	if pwsh, err := exec.LookPath("pwsh.exe"); err == nil {
		return pwsh
	}
	return "powershell.exe"
}

// gitBash finds Git for Windows' bash. The bash.exe in System32 is WSL's,
// which runs in a Linux file system, so it is skipped.
func gitBash() string {
	candidates := []string{
		filepath.Join(os.Getenv("ProgramFiles"), "Git", "bin", "bash.exe"),
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "Git", "bin", "bash.exe"),
		filepath.Join(os.Getenv("LocalAppData"), "Programs", "Git", "bin", "bash.exe"),
	}
	if onPath, err := exec.LookPath("bash.exe"); err == nil && !strings.Contains(strings.ToLower(onPath), `\windows\system32`) {
		candidates = append([]string{onPath}, candidates...)
	}
	for _, path := range candidates {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
	}
	return ""
}
