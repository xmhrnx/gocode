package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"gocode/internal/tools"
)

// stubCaller returns a fixed result or error.
type stubCaller struct {
	result callToolResult
	err    error
}

func (s stubCaller) callTool(context.Context, string, json.RawMessage) (callToolResult, error) {
	return s.result, s.err
}

func TestServerToolSpecDefaultsSchema(t *testing.T) {
	tool := &serverTool{server: "s", info: toolInfo{Name: "t", Description: "does t"}, name: "mcp__s__t"}
	spec := tool.Spec()
	if spec.Name != "mcp__s__t" || spec.Description != "does t" || string(spec.Parameters) != string(emptyObjectSchema) {
		t.Errorf("spec = %+v", spec)
	}
}

func TestServerToolSummary(t *testing.T) {
	tool := &serverTool{server: "github", info: toolInfo{Name: "search"}}
	if got := tool.Summary(json.RawMessage(`{ "query": "x" }`)); got != `github: search {"query":"x"}` {
		t.Errorf("summary = %q", got)
	}
	if got := tool.Summary(json.RawMessage(`{}`)); got != "github: search" {
		t.Errorf("summary without args = %q", got)
	}
	long := tool.Summary(json.RawMessage(`{"q":"` + strings.Repeat("z", 200) + `"}`))
	if len(long) > len("github: search ")+summaryArgsLimit+len("…") {
		t.Errorf("summary not truncated: %d bytes", len(long))
	}
}

func TestServerToolRun(t *testing.T) {
	big := strings.Repeat("a", 100)
	tool := &serverTool{caller: stubCaller{result: callToolResult{Content: []contentItem{{Type: "text", Text: big}}, IsError: true}}}
	result := tool.Run(context.Background(), tools.Env{MaxOutputBytes: 40}, nil)
	if !result.IsError || !strings.Contains(result.Output, "truncated") {
		t.Errorf("result = %+v", result)
	}
	failing := &serverTool{caller: stubCaller{err: errors.New("down")}}
	if result := failing.Run(context.Background(), tools.Env{}, nil); !result.IsError || result.Output != "down" {
		t.Errorf("error result = %+v", result)
	}
}
