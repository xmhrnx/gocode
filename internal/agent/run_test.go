package agent

import (
	"context"
	"testing"

	"gocode/internal/llm"
	"gocode/internal/settings"
	"gocode/internal/tools"
)

func newTestConfig(client llm.Client, access tools.Access, ask Asker) Config {
	return Config{
		Client: client, Model: "m", Tools: tools.NewRegistry(echoTool{access: access}),
		Permissions: NewPermissions(ModeAsk, settings.Default().Agent), Ask: ask, MaxSteps: 5,
	}
}

func TestRunExecutesToolsUntilModelStops(t *testing.T) {
	client := &scriptedClient{replies: [][]llm.Event{
		{{Kind: llm.EventText, Text: "Let me "}, {Kind: llm.EventText, Text: "check."}, toolCallEvent("1", "pong")},
		{{Kind: llm.EventText, Text: "Done."}},
	}}
	added, err := Run(context.Background(), newTestConfig(client, tools.AccessRead, nil), nil, func(Event) {})
	if err != nil {
		t.Fatal(err)
	}
	if len(added) != 3 {
		t.Fatalf("want assistant, tool result, assistant; got %d messages", len(added))
	}
	if got := added[0].Text(); got != "Let me check." {
		t.Errorf("text parts not merged: %q", got)
	}
	result := added[1].Parts[0].ToolResult
	if result.Content != "pong" || result.CallID != "1" {
		t.Errorf("tool result = %+v", result)
	}
	if len(client.requests[1].Messages) != 2 {
		t.Errorf("second request should carry the tool round trip")
	}
}

func TestRunReportsDenialWithFeedback(t *testing.T) {
	client := &scriptedClient{replies: [][]llm.Event{
		{toolCallEvent("1", "x")},
		{{Kind: llm.EventText, Text: "OK."}},
	}}
	deny := func(context.Context, PermissionRequest) PermissionAnswer {
		return PermissionAnswer{Decision: DecisionDeny, Feedback: "use tests instead"}
	}
	added, err := Run(context.Background(), newTestConfig(client, tools.AccessExecute, deny), nil, func(Event) {})
	if err != nil {
		t.Fatal(err)
	}
	result := added[1].Parts[0].ToolResult
	if !result.IsError || result.Content != "The user denied this tool call. Their feedback: use tests instead" {
		t.Errorf("denial result = %+v", result)
	}
}

func TestReadOnlyModeDeniesExecution(t *testing.T) {
	perms := NewPermissions(ModeReadOnly, settings.Default().Agent)
	if got := perms.Policy(echoTool{access: tools.AccessExecute}); got != settings.PolicyDeny {
		t.Fatalf("policy = %s", got)
	}
	if got := perms.Policy(echoTool{access: tools.AccessRead}); got != settings.PolicyAllow {
		t.Fatalf("read policy = %s", got)
	}
}
