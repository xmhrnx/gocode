package shell

import (
	"os/exec"
	"strings"
	"testing"
)

// Tabs, quotes, backslashes and heredocs must reach the shell unchanged.
func TestScriptArgsPassScriptsUnchanged(t *testing.T) {
	script := "cat <<'EOF'\n\tPublic \"quoted\" \\back\\slash\nEOF"
	args, cleanup, err := ScriptArgs("/bin/sh", script)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	out, err := exec.Command("/bin/sh", args...).Output()
	if err != nil {
		t.Fatal(err)
	}
	if want := "\tPublic \"quoted\" \\back\\slash\n"; string(out) != want {
		t.Fatalf("got %q, want %q", out, want)
	}
	if strings.Contains(string(out), "/Public") {
		t.Fatal("path conversion")
	}
}
