package tools

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestShellArgs(t *testing.T) {
	tests := []struct {
		shell string
		want  []string
	}{
		{"/bin/bash", []string{"-c", "x"}},
		{"powershell.exe", []string{"-NoProfile", "-NonInteractive", "-Command", "x"}},
		{"pwsh", []string{"-NoProfile", "-NonInteractive", "-Command", "x"}},
		{"cmd.exe", []string{"/C", "x"}},
	}
	for _, tt := range tests {
		if got := shellArgs(tt.shell, "x"); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("shellArgs(%q) = %v, want %v", tt.shell, got, tt.want)
		}
	}
}

func TestCommandTimeout(t *testing.T) {
	tests := []struct {
		envSeconds, requested int
		want                  time.Duration
	}{
		{0, 0, 120 * time.Second},
		{30, 0, 30 * time.Second},
		{30, 5, 5 * time.Second},
		{0, 9999, 600 * time.Second},
	}
	for _, tt := range tests {
		if got := commandTimeout(Env{CommandTimeoutSeconds: tt.envSeconds}, tt.requested); got != tt.want {
			t.Errorf("commandTimeout(%d, %d) = %v, want %v", tt.envSeconds, tt.requested, got, tt.want)
		}
	}
}

func TestBashSummary(t *testing.T) {
	withDescription := json.RawMessage(`{"command":"go test ./...","description":"Run the tests"}`)
	if got := (Bash{}).Summary(withDescription); got != "Run the tests" {
		t.Errorf("got %q", got)
	}
	if got := (Bash{}).Summary(json.RawMessage(`{"command":"go test ./...\necho done"}`)); got != "Run go test ./... …" {
		t.Errorf("got %q", got)
	}
}
