package prompts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildExpandsVarsAndAppendsRules(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("use tabs"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := Build("wd={{workdir}}", "be brief", Vars{WorkDir: dir}, true)
	for _, want := range []string{"wd=" + dir, "be brief", "use tabs"} {
		if !strings.Contains(got, want) {
			t.Errorf("prompt lacks %q:\n%s", want, got)
		}
	}
}

func TestOrFallsBackOnBlank(t *testing.T) {
	if Or("  ", "x") != "x" || Or("y", "x") != "y" {
		t.Fatal("Or misbehaves")
	}
}
