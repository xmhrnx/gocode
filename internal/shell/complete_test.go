package shell

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestCompletePaths(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"main.go", "makefile.mk", "my notes.txt", ".env"} {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"internal", "internal/agent", "internal/agents"} {
		if err := os.Mkdir(filepath.Join(dir, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	cases := []struct {
		line, want string
		matches    []string
	}{
		{"cat mai", "cat main.go ", nil},
		{"cat ma", "cat ma", []string{"main.go", "makefile.mk"}},
		{"cd int", "cd internal/", nil},
		{"cd internal/a", "cd internal/agent", []string{"agent/", "agents/"}},
		{"cat .e", "cat .env ", nil},
		{"cat zz", "cat zz", nil},
	}
	if runtime.GOOS != "windows" {
		cases = append(cases, struct {
			line, want string
			matches    []string
		}{"cat my", `cat my\ notes.txt `, nil})
	}
	for _, c := range cases {
		got := Complete(c.line, dir)
		if got.Line != c.want || !reflect.DeepEqual(got.Matches, c.matches) {
			t.Errorf("Complete(%q) = %q %v, want %q %v", c.line, got.Line, got.Matches, c.want, c.matches)
		}
	}
}

func TestCompleteCommands(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("executables are found by extension on Windows")
	}
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "gocodetool"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	if got := Complete("gocodet", t.TempDir()).Line; got != "gocodetool " {
		t.Errorf("command completion = %q", got)
	}
	if got := Complete("ex", t.TempDir()); got.Line != "ex" || len(got.Matches) != 2 {
		t.Errorf("builtins = %q %v, want exit and export", got.Line, got.Matches)
	}
}
