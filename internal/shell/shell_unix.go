//go:build !windows

package shell

import (
	"os"
)

// Default returns the shell commands run in: $SHELL, falling back to bash.
func Default() string {
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell
	}
	return "/bin/bash"
}
