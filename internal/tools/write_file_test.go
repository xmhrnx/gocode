package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteFileCreatesWithParents(t *testing.T) {
	env := newTestEnv(t, nil)
	args := writeFileArgs{Path: "sub/dir/new.txt", Content: "hello\n"}
	raw, _ := json.Marshal(args)
	preview, err := WriteFile{}.Preview(env, raw)
	if err != nil || !strings.Contains(preview, "--- /dev/null\n+++ b/sub/dir/new.txt\n") {
		t.Fatalf("preview %q, err %v", preview, err)
	}
	res := runTool(t, WriteFile{}, env, args)
	if res.IsError || !strings.HasPrefix(res.Output, "Created sub/dir/new.txt") || res.Display != preview {
		t.Fatalf("got %+v", res)
	}
	if got := readTestFile(t, env, "sub/dir/new.txt"); got != "hello\n" {
		t.Errorf("content %q", got)
	}
}

func TestWriteFileOverwriteKeepsMode(t *testing.T) {
	env := newTestEnv(t, map[string]string{"run.sh": "echo old\n"})
	path := filepath.Join(env.WorkDir, "run.sh")
	if err := os.Chmod(path, 0o755); err != nil {
		t.Fatal(err)
	}
	res := runTool(t, WriteFile{}, env, writeFileArgs{Path: "run.sh", Content: "echo new\n"})
	if res.IsError || !strings.Contains(res.Display, "-echo old\n+echo new\n") {
		t.Fatalf("got %+v", res)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Errorf("mode %v, want 0755", info.Mode().Perm())
	}
}

func TestWriteFileRefusesDirectory(t *testing.T) {
	env := newTestEnv(t, map[string]string{"d/x": ""})
	if res := runTool(t, WriteFile{}, env, writeFileArgs{Path: "d", Content: "x"}); !res.IsError {
		t.Errorf("expected error, got %+v", res)
	}
}
