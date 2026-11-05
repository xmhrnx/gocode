package tools

import (
	"path/filepath"
	"strings"
)

// shellArgs returns the arguments that make shell run command, according
// to the shell's calling convention.
func shellArgs(shell, command string) []string {
	name := strings.ToLower(strings.TrimSuffix(filepath.Base(shell), filepath.Ext(shell)))
	switch name {
	case "powershell", "pwsh":
		return []string{"-NoProfile", "-NonInteractive", "-Command", command}
	case "cmd":
		return []string{"/C", command}
	}
	return []string{"-c", command}
}
