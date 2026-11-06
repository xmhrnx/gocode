package shell

import (
	"strings"
	"testing"
)

func TestWrapScript(t *testing.T) {
	posix := wrapScript("/bin/zsh", "make", "/tmp/x")
	if !strings.HasPrefix(posix, "trap ") || !strings.HasSuffix(posix, " EXIT\nmake") {
		t.Errorf("posix wrapper %q", posix)
	}
	ps := wrapScript(`powershell.exe`, "dir", `C:\Temp\x`)
	if !strings.Contains(ps, `WriteAllText('C:\Temp\x'`) || !strings.Contains(ps, "\ndir\n") {
		t.Errorf("powershell wrapper %q", ps)
	}
}

func TestKindOf(t *testing.T) {
	cases := map[string]Kind{"/bin/bash": KindPOSIX, `C:\Git\bin\bash.exe`: KindPOSIX,
		"powershell.exe": KindPowerShell, "pwsh": KindPowerShell, "cmd.exe": KindCmd}
	for program, want := range cases {
		if got := KindOf(program); got != want {
			t.Errorf("KindOf(%q) = %v, want %v", program, got, want)
		}
	}
}
