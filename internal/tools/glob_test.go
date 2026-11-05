package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCompileGlob(t *testing.T) {
	tests := []struct {
		pattern, path string
		want          bool
	}{
		{"*.go", "main.go", true},
		{"*.go", "cmd/main.go", false},
		{"**/*.go", "main.go", true},
		{"**/*.go", "a/b/main.go", true},
		{"src/**/test_*.py", "src/test_a.py", true},
		{"src/**/test_*.py", "src/x/y/test_a.py", true},
		{"src/**/test_*.py", "lib/test_a.py", false},
		{"*.{ts,tsx}", "a.tsx", true},
		{"*.{ts,tsx}", "a.js", false},
		{"file?.txt", "file1.txt", true},
		{"[!a]*", "abc", false},
		{"[a-c]*", "bcd", true},
		{"a.b", "axb", false},
		{"docs/**", "docs/x/y.md", true},
	}
	for _, tt := range tests {
		re, err := compileGlob(tt.pattern)
		if err != nil {
			t.Fatalf("compileGlob(%q): %v", tt.pattern, err)
		}
		if got := re.MatchString(tt.path); got != tt.want {
			t.Errorf("%q matches %q = %v, want %v (re %s)", tt.pattern, tt.path, got, tt.want, re)
		}
	}
	for _, bad := range []string{"[abc", "{a,b"} {
		if _, err := compileGlob(bad); err == nil {
			t.Errorf("compileGlob(%q) should fail", bad)
		}
	}
}

func TestGlobSortsNewestFirst(t *testing.T) {
	env := newTestEnv(t, map[string]string{
		"old.go":            "",
		"pkg/new.go":        "",
		"pkg/readme.md":     "",
		"node_modules/x.go": "",
	})
	past := time.Now().Add(-time.Hour)
	if err := os.Chtimes(filepath.Join(env.WorkDir, "old.go"), past, past); err != nil {
		t.Fatal(err)
	}
	res := runTool(t, Glob{}, env, globArgs{Pattern: "**/*.go"})
	if res.IsError || res.Output != "pkg/new.go\nold.go" {
		t.Errorf("got %q", res.Output)
	}
	sub := runTool(t, Glob{}, env, globArgs{Pattern: "*.md", Path: "pkg"})
	if sub.Output != "pkg/readme.md" {
		t.Errorf("got %q", sub.Output)
	}
	none := runTool(t, Glob{}, env, globArgs{Pattern: "*.rs"})
	if !strings.Contains(none.Output, "No files found") {
		t.Errorf("got %q", none.Output)
	}
}
