package tools

import (
	"strings"
	"testing"
)

func grepTestEnv(t *testing.T) Env {
	return newTestEnv(t, map[string]string{
		"a.go":              "package a\n\nfunc Hello() {}\nfunc other() {}\n",
		"sub/b.go":          "package b\n// hello there\n",
		"notes.txt":         "Hello notes\n",
		"bin.dat":           "Hello\x00",
		"node_modules/x.go": "func Hello() {}\n",
	})
}

func TestGrep(t *testing.T) {
	env := grepTestEnv(t)
	tests := []struct {
		name string
		args grepArgs
		want string
	}{
		{"content", grepArgs{Pattern: "Hello"}, "a.go:3: func Hello() {}\nnotes.txt:1: Hello notes"},
		{"case insensitive include", grepArgs{Pattern: "hello", CaseInsensitive: true, Include: "*.go"},
			"a.go:3: func Hello() {}\nsub/b.go:2: // hello there"},
		{"files", grepArgs{Pattern: "^package", OutputMode: "files"}, "a.go\nsub/b.go"},
		{"single file", grepArgs{Pattern: "package", Path: "sub/b.go"}, "sub/b.go:1: package b"},
		{"context", grepArgs{Pattern: "Hello", Include: "*.go", Context: 1},
			"a.go-2- \na.go:3: func Hello() {}\na.go-4- func other() {}"},
		{"none", grepArgs{Pattern: "zzz"}, "No matches found."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := runTool(t, Grep{}, env, tt.args)
			if res.IsError || res.Output != tt.want {
				t.Errorf("got %q (error %v), want %q", res.Output, res.IsError, tt.want)
			}
		})
	}
}

func TestGrepContextSeparatesGroups(t *testing.T) {
	env := newTestEnv(t, map[string]string{"f.txt": "x\n1\n2\n3\n4\nx\n"})
	res := runTool(t, Grep{}, env, grepArgs{Pattern: "x", Context: 1})
	want := "f.txt:1: x\nf.txt-2- 1\n--\nf.txt-5- 4\nf.txt:6: x"
	if res.Output != want {
		t.Errorf("got %q, want %q", res.Output, want)
	}
}

func TestGrepErrorsAndCap(t *testing.T) {
	env := newTestEnv(t, map[string]string{"many.txt": strings.Repeat("hit\n", maxGrepMatches+5)})
	for _, args := range []grepArgs{{Pattern: "("}, {Pattern: ""}, {Pattern: "a", OutputMode: "lines"}, {Pattern: "a", Context: 9}} {
		if res := runTool(t, Grep{}, env, args); !res.IsError {
			t.Errorf("args %+v should fail, got %q", args, res.Output)
		}
	}
	res := runTool(t, Grep{}, env, grepArgs{Pattern: "hit"})
	if strings.Count(res.Output, "many.txt:") != maxGrepMatches || !strings.Contains(res.Output, "stopped after 300 matches") {
		t.Errorf("cap not applied: %d matches", strings.Count(res.Output, "many.txt:"))
	}
}
