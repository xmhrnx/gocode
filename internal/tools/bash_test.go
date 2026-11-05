//go:build !windows

package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestBash(t *testing.T) {
	env := newTestEnv(t, map[string]string{"marker.txt": ""})
	env.Shell = "/bin/sh"
	tests := []struct {
		name    string
		command string
		want    string
		wantErr bool
	}{
		{"runs in workdir", "ls", "marker.txt\n", false},
		{"combines stderr", "echo out; echo err >&2", "out\nerr\n", false},
		{"no output", "true", "(no output)", false},
		{"exit code", "echo bad; exit 3", "bad\n\nexit code: 3", true},
		{"stdin closed", "cat", "(no output)", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := runTool(t, Bash{}, env, bashArgs{Command: tt.command})
			if res.IsError != tt.wantErr || res.Output != tt.want {
				t.Errorf("got %q (error %v), want %q", res.Output, res.IsError, tt.want)
			}
		})
	}
}

func TestBashTimeoutKillsProcessGroup(t *testing.T) {
	env := Env{WorkDir: t.TempDir(), Shell: "/bin/sh"}
	start := time.Now()
	// The background sleep shares the pipe; only killing the whole group
	// lets the call return promptly.
	res := runTool(t, Bash{}, env, bashArgs{Command: "sleep 30 & sleep 30", TimeoutSeconds: 1})
	if !res.IsError || !strings.Contains(res.Output, "timed out after 1s") {
		t.Errorf("got %+v", res)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("took %v; process group not killed", elapsed)
	}
}

func TestBashCancel(t *testing.T) {
	env := Env{WorkDir: t.TempDir(), Shell: "/bin/sh"}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	raw, _ := json.Marshal(bashArgs{Command: "sleep 30"})
	res := Bash{}.Run(ctx, env, raw)
	if !res.IsError || !strings.Contains(res.Output, "cancelled") {
		t.Errorf("got %+v", res)
	}
}

func TestBashTruncatesKeepingTail(t *testing.T) {
	env := Env{WorkDir: t.TempDir(), Shell: "/bin/sh", MaxOutputBytes: 200}
	res := runTool(t, Bash{}, env, bashArgs{Command: "seq 1 1000; echo FINAL"})
	if !strings.HasPrefix(res.Output, "1\n2\n") || !strings.HasSuffix(res.Output, "FINAL\n") || !strings.Contains(res.Output, "[truncated:") {
		t.Errorf("got %q", res.Output)
	}
}
