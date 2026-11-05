package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

// connectFake starts the fake server and closes it when the test ends.
func connectFake(t *testing.T) *client {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	c, err := connect(ctx, fakeServerConfig("fake"), serverEvents{})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() {
		if err := c.close(); err != nil {
			t.Errorf("close: %v", err)
		}
	})
	return c
}

func callFake(t *testing.T, c *client, name string, args string) callToolResult {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, err := c.callTool(ctx, name, json.RawMessage(args))
	if err != nil {
		t.Fatalf("call %s: %v", name, err)
	}
	return result
}

func TestStdioListToolsFollowsPages(t *testing.T) {
	c := connectFake(t)
	list, err := c.listTools(context.Background())
	if err != nil {
		t.Fatalf("listTools: %v", err)
	}
	if len(list) != 2 || list[0].Name != "echo" || list[1].Name != "fail" {
		t.Fatalf("tools = %+v, want echo and fail", list)
	}
	if list[1].Annotations == nil || !list[1].Annotations.ReadOnlyHint {
		t.Errorf("fail tool lost its readOnlyHint")
	}
}

func TestStdioCallToolText(t *testing.T) {
	c := connectFake(t)
	result := callFake(t, c, "echo", `{"text":"hello"}`)
	if want := "hello\n[image: image/png, 2 KB]"; result.text() != want || result.IsError {
		t.Errorf("echo = %q (error %v), want %q", result.text(), result.IsError, want)
	}
	failed := callFake(t, c, "fail", `{}`)
	if !failed.IsError || failed.text() != "boom" {
		t.Errorf("fail = %q (error %v), want boom as error", failed.text(), failed.IsError)
	}
}

func TestStdioAnswersServerRequests(t *testing.T) {
	c := connectFake(t)
	result := callFake(t, c, "ask_client", `{}`)
	if want := "ping:{} sampling:-32601"; result.text() != want {
		t.Errorf("server saw %q, want %q", result.text(), want)
	}
}

func TestStdioUnknownMethodIsRPCError(t *testing.T) {
	c := connectFake(t)
	err := c.request(context.Background(), "resources/list", nil, nil)
	var rpcErr *RPCError
	if !errors.As(err, &rpcErr) || rpcErr.Code != codeMethodNotFound {
		t.Fatalf("err = %v, want RPC error %d", err, codeMethodNotFound)
	}
}

func TestStdioCallHonoursDeadline(t *testing.T) {
	c := connectFake(t)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if _, err := c.callTool(ctx, "slow", json.RawMessage(`{}`)); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want deadline exceeded", err)
	}
	// The connection must still work after an abandoned call.
	callFake(t, c, "echo", `{"text":"still here"}`)
}

func TestStdioCrashReportsStderr(t *testing.T) {
	c := connectFake(t)
	_, err := c.callTool(context.Background(), "crash", json.RawMessage(`{}`))
	if err == nil || !strings.Contains(err.Error(), "fake server crashed") {
		t.Fatalf("err = %v, want the server's stderr", err)
	}
	if _, err := c.listTools(context.Background()); err == nil {
		t.Error("call after crash succeeded")
	}
}
