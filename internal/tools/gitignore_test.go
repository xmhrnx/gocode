package tools

import (
	"path/filepath"
	"testing"
)

func TestGitignorePatterns(t *testing.T) {
	tests := []struct {
		line  string
		path  string
		isDir bool
		want  bool
	}{
		{"*.log", "a/b/debug.log", false, true},
		{"*.log", "a/log.txt", false, false},
		{"build/", "sub/build", true, true},
		{"build/", "sub/build", false, false},
		{"/out", "out", true, true},
		{"/out", "sub/out", true, false},
		{"docs/gen", "docs/gen", true, true},
		{"docs/gen", "x/docs/gen", true, false},
		{"**/tmp", "a/b/tmp", true, true},
		{"secret.txt", "deep/secret.txt", false, true},
	}
	for _, tt := range tests {
		p, ok := parseGitignoreLine(tt.line)
		if !ok {
			t.Fatalf("pattern %q not parsed", tt.line)
		}
		if got := p.matches(tt.path, tt.isDir); got != tt.want {
			t.Errorf("%q matches %q (dir=%v) = %v, want %v", tt.line, tt.path, tt.isDir, got, tt.want)
		}
	}
}

func TestGitignoreSkipsCommentsAndNegations(t *testing.T) {
	for _, line := range []string{"", "  ", "# comment", "!keep.log", "/"} {
		if _, ok := parseGitignoreLine(line); ok {
			t.Errorf("line %q should be skipped", line)
		}
	}
}

func TestIgnoreRules(t *testing.T) {
	env := newTestEnv(t, map[string]string{".gitignore": "# generated\n*.tmp\ncoverage/\n"})
	rules := loadIgnoreRules(env.WorkDir)
	tests := []struct {
		rel   string
		isDir bool
		want  bool
	}{
		{"node_modules", true, true},
		{"src/.git", true, true},
		{"x.tmp", false, true},
		{"coverage", true, true},
		{"vendor", true, false},
		{"main.go", false, false},
	}
	for _, tt := range tests {
		if got := rules.skip(filepath.Join(env.WorkDir, tt.rel), tt.isDir); got != tt.want {
			t.Errorf("skip(%q) = %v, want %v", tt.rel, got, tt.want)
		}
	}
}
