package mcp

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"testing"
	"time"

	"gocode/internal/settings"
	"gocode/internal/tools"
)

// waitForStatus polls the manager until ok accepts the named server's
// status, failing the test after a deadline.
func waitForStatus(t *testing.T, m *Manager, name string, ok func(ServerStatus) bool) ServerStatus {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		for _, status := range m.Status() {
			if status.Name == name && ok(status) {
				return status
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("server %s never reached the expected state; statuses: %+v", name, m.Status())
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func isConnected(status ServerStatus) bool { return status.State == StateConnected }

// findTool returns the manager's tool with name.
func findTool(t *testing.T, m *Manager, name string) tools.Tool {
	t.Helper()
	for _, tool := range m.Tools() {
		if tool.Spec().Name == name {
			return tool
		}
	}
	t.Fatalf("no tool %s", name)
	return nil
}

// unlistedTool wraps a tool the fake server serves but does not list.
func unlistedTool(m *Manager, server, name string) *serverTool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return &serverTool{caller: m.servers[server].client, server: server, info: toolInfo{Name: name}, name: name}
}

func TestManagerConnectsAndExposesTools(t *testing.T) {
	m := NewManager()
	defer m.Close()
	changes := make(chan struct{}, 100)
	m.OnChange(func() { changes <- struct{}{} })
	disabled := fakeServerConfig("off")
	disabled.Enabled = false
	m.Sync([]settings.MCPServer{fakeServerConfig("fake"), disabled})

	status := waitForStatus(t, m, "fake", isConnected)
	if !slices.Equal(status.ToolNames, []string{"echo", "fail"}) {
		t.Errorf("tool names = %v", status.ToolNames)
	}
	if got := m.Status()[1].State; got != StateDisabled {
		t.Errorf("disabled server state = %s", got)
	}
	if len(changes) == 0 {
		t.Error("OnChange was not called")
	}
	echo := findTool(t, m, "mcp__fake__echo")
	if echo.Access() != tools.AccessExecute || findTool(t, m, "mcp__fake__fail").Access() != tools.AccessRead {
		t.Error("access does not follow readOnlyHint")
	}
	result := echo.Run(context.Background(), tools.Env{}, json.RawMessage(`{"text":"hi"}`))
	if result.IsError || !strings.HasPrefix(result.Output, "hi\n") {
		t.Errorf("echo result = %+v", result)
	}

	m.Sync([]settings.MCPServer{disabled})
	if len(m.Tools()) != 0 || len(m.Status()) != 1 {
		t.Errorf("removed server still listed: %+v", m.Status())
	}
}

func TestManagerRelistsToolsOnListChanged(t *testing.T) {
	m := NewManager()
	defer m.Close()
	m.Sync([]settings.MCPServer{fakeServerConfig("fake")})
	waitForStatus(t, m, "fake", isConnected)
	unlistedTool(m, "fake", "change_tools").Run(context.Background(), tools.Env{}, nil)
	waitForStatus(t, m, "fake", func(s ServerStatus) bool { return slices.Contains(s.ToolNames, "added") })
	findTool(t, m, "mcp__fake__added")
}

func TestManagerReportsCrashAndReconnects(t *testing.T) {
	m := NewManager()
	defer m.Close()
	m.Sync([]settings.MCPServer{fakeServerConfig("fake")})
	waitForStatus(t, m, "fake", isConnected)
	unlistedTool(m, "fake", "crash").Run(context.Background(), tools.Env{}, nil)
	failed := waitForStatus(t, m, "fake", func(s ServerStatus) bool { return s.State == StateFailed })
	if !strings.Contains(failed.Err, "fake server crashed") || len(m.Tools()) != 0 {
		t.Errorf("after crash: %+v, %d tools", failed, len(m.Tools()))
	}
	m.Reconnect("fake")
	waitForStatus(t, m, "fake", isConnected)
}

func TestManagerReportsFailedStart(t *testing.T) {
	m := NewManager()
	defer m.Close()
	m.Sync([]settings.MCPServer{{Name: "missing", Command: "/nonexistent/mcp-server", Enabled: true}})
	failed := waitForStatus(t, m, "missing", func(s ServerStatus) bool { return s.State == StateFailed })
	if failed.Err == "" {
		t.Error("failure has no explanation")
	}
}
