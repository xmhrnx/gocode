package shell

import (
	"path/filepath"
	"strings"
)

// Kind is the calling convention a shell program follows.
type Kind int

const (
	// KindPOSIX is bash, zsh, sh and friends, Git Bash included.
	KindPOSIX Kind = iota
	// KindPowerShell is Windows PowerShell or PowerShell 7 (pwsh).
	KindPowerShell
	// KindCmd is the Windows command prompt.
	KindCmd
)

// KindOf tells which calling convention program follows, by its name.
func KindOf(program string) Kind {
	switch strings.ToLower(strings.TrimSuffix(filepath.Base(program), filepath.Ext(program))) {
	case "powershell", "pwsh":
		return KindPowerShell
	case "cmd":
		return KindCmd
	}
	return KindPOSIX
}
