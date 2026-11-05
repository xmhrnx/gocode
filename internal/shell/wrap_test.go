package shell

import (
	"strings"
	"testing"
)

func TestWrapCommand(t *testing.T) {
	posix := wrapCommand("/bin/zsh", "make", "/tmp/x")
	if posix[0] != "-c" || !strings.HasPrefix(posix[1], "trap ") || !strings.HasSuffix(posix[1], " EXIT\nmake") {
		t.Errorf("posix wrapper %q", posix)
	}
	ps := wrapCommand(`powershell.exe`, "dir", `C:\Temp\x`)
	if ps[2] != "-Command" || !strings.Contains(ps[3], `WriteAllText('C:\Temp\x'`) || !strings.Contains(ps[3], "\ndir\n") {
		t.Errorf("powershell wrapper %q", ps)
	}
}
