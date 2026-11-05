package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// newTestEnv returns an Env rooted in a fresh directory holding files.
func newTestEnv(t *testing.T, files map[string]string) Env {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return Env{WorkDir: dir}
}

// runTool calls tool with args encoded as JSON.
func runTool(t *testing.T, tool Tool, env Env, args any) Result {
	t.Helper()
	raw, err := json.Marshal(args)
	if err != nil {
		t.Fatal(err)
	}
	return tool.Run(context.Background(), env, raw)
}

// readTestFile returns a file's content relative to env's root.
func readTestFile(t *testing.T, env Env, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(env.WorkDir, filepath.FromSlash(name)))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
