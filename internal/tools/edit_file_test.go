package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEditFile(t *testing.T) {
	tests := []struct {
		name    string
		initial string
		args    editFileArgs
		want    string // file content, or error substring
		wantErr bool
	}{
		{"single", "a\nb\nc\n", editFileArgs{OldString: "b", NewString: "B"}, "a\nB\nc\n", false},
		{"not found", "a\n", editFileArgs{OldString: "z", NewString: "y"}, "not found", true},
		{"ambiguous", "x x\n", editFileArgs{OldString: "x", NewString: "y"}, "matches 2 places", true},
		{"replace all", "x x\n", editFileArgs{OldString: "x", NewString: "y", ReplaceAll: true}, "y y\n", false},
		{"identical", "x\n", editFileArgs{OldString: "x", NewString: "x"}, "identical", true},
		{"empty old on content", "x\n", editFileArgs{OldString: "", NewString: "y"}, "already has content", true},
		{"crlf file", "one\r\ntwo\r\nthree\r\n", editFileArgs{OldString: "one\ntwo\n", NewString: "1\n2\n"}, "1\r\n2\r\nthree\r\n", false},
		{"crlf exact", "one\r\ntwo\r\n", editFileArgs{OldString: "one\r\n", NewString: "1\r\n"}, "1\r\ntwo\r\n", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newTestEnv(t, map[string]string{"f.txt": tt.initial})
			tt.args.Path = "f.txt"
			res := runTool(t, EditFile{}, env, tt.args)
			if res.IsError != tt.wantErr {
				t.Fatalf("IsError = %v: %s", res.IsError, res.Output)
			}
			if tt.wantErr {
				if !strings.Contains(res.Output, tt.want) {
					t.Errorf("error %q lacks %q", res.Output, tt.want)
				}
				return
			}
			if got := readTestFile(t, env, "f.txt"); got != tt.want {
				t.Errorf("content %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEditFileCreatesAndPreviews(t *testing.T) {
	env := newTestEnv(t, nil)
	args := editFileArgs{Path: "new/file.go", OldString: "", NewString: "package x\n"}
	res := runTool(t, EditFile{}, env, args)
	if res.IsError || res.Output != "Created new/file.go" || !strings.Contains(res.Display, "+package x") {
		t.Fatalf("got %+v", res)
	}
	missing := runTool(t, EditFile{}, env, editFileArgs{Path: "gone.go", OldString: "a", NewString: "b"})
	if !missing.IsError || !strings.Contains(missing.Output, "not found") {
		t.Errorf("got %+v", missing)
	}
}

func TestEditFilePreviewDoesNotWrite(t *testing.T) {
	env := newTestEnv(t, map[string]string{"f.txt": "1\n2\n3\n4\n5\n"})
	if err := os.Chmod(filepath.Join(env.WorkDir, "f.txt"), 0o600); err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(editFileArgs{Path: "f.txt", OldString: "3\n", NewString: "three\n"})
	preview, err := EditFile{}.Preview(env, raw)
	if err != nil || !strings.Contains(preview, "@@ -1,5 +1,5 @@\n 1\n 2\n-3\n+three\n 4\n 5\n") {
		t.Fatalf("preview %q, err %v", preview, err)
	}
	if got := readTestFile(t, env, "f.txt"); got != "1\n2\n3\n4\n5\n" {
		t.Errorf("preview changed the file: %q", got)
	}
	if res := (EditFile{}).Run(t.Context(), env, raw); res.IsError || res.Display != preview {
		t.Fatalf("got %+v", res)
	}
	info, _ := os.Stat(filepath.Join(env.WorkDir, "f.txt"))
	if info.Mode().Perm() != 0o600 {
		t.Errorf("mode %v, want 0600", info.Mode().Perm())
	}
}
