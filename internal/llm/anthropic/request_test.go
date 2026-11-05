package anthropic

import (
	"context"
	"encoding/json"
	"testing"

	"gocode/internal/llm"
)

func TestBuildRequestConvertsHistory(t *testing.T) {
	temperature := 0.7
	call := llm.ToolCall{ID: "toolu_1", Name: "read", Arguments: json.RawMessage(`{"path":"a.go"}`)}
	history := []llm.Message{
		{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart("look"), llm.TextPart(""), llm.ImagePart("image/png", []byte{1, 2, 3})}},
		{Role: llm.RoleAssistant, Parts: []llm.Part{
			{Kind: llm.PartReasoning, Text: "plan", Signature: "sig"},
			{Kind: llm.PartReasoning, Signature: "opaque", Redacted: true},
			{Kind: llm.PartReasoning, Text: "cancelled, unsigned"},
			llm.TextPart("Reading."),
			{Kind: llm.PartToolCall, ToolCall: &call},
		}},
		{Role: llm.RoleUser, Parts: []llm.Part{
			llm.TextPart("note"),
			llm.ToolResultPart(llm.ToolResult{CallID: "toolu_1", Content: "body", IsError: true}),
		}},
		{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart("more")}},
	}
	got := mustJSON(t, buildRequest(llm.Request{
		Model: "claude-opus-5-5", System: "sys", Messages: history, Temperature: &temperature,
		ReasoningEffort: "high", Tools: []llm.ToolSpec{{Name: "read", Description: "Read"}},
	}))
	want := `{"model":"claude-opus-5-5","max_tokens":32000,"system":"sys","messages":[` +
		`{"role":"user","content":[{"type":"text","text":"look"},{"type":"image","source":{"type":"base64","media_type":"image/png","data":"AQID"}}]},` +
		`{"role":"assistant","content":[{"type":"thinking","thinking":"plan","signature":"sig"},{"type":"redacted_thinking","data":"opaque"},` +
		`{"type":"text","text":"Reading."},{"type":"tool_use","id":"toolu_1","name":"read","input":{"path":"a.go"}}]},` +
		`{"role":"user","content":[{"type":"tool_result","tool_use_id":"toolu_1","content":"body","is_error":true},{"type":"text","text":"note"},{"type":"text","text":"more"}]}],` +
		`"tools":[{"name":"read","description":"Read","input_schema":{"type":"object","properties":{}}}],"stream":true,` +
		`"thinking":{"type":"adaptive","display":"summarized"},"output_config":{"effort":"high"},"cache_control":{"type":"ephemeral"}}`
	if got != want {
		t.Errorf("request JSON\n got: %s\nwant: %s", got, want)
	}
}

func TestBuildRequestWithoutReasoningEffort(t *testing.T) {
	got := mustJSON(t, buildRequest(llm.Request{Model: "m", MaxTokens: 100, Messages: []llm.Message{
		{Role: llm.RoleAssistant, Parts: []llm.Part{{Kind: llm.PartToolCall, ToolCall: &llm.ToolCall{ID: "t", Name: "n", Arguments: json.RawMessage(`"bad"`)}}}},
	}}))
	want := `{"model":"m","max_tokens":100,"messages":[{"role":"assistant","content":[{"type":"tool_use","id":"t","name":"n","input":{}}]}],` +
		`"stream":true,"cache_control":{"type":"ephemeral"}}`
	if got != want {
		t.Errorf("request JSON\n got: %s\nwant: %s", got, want)
	}
}

func TestListModels(t *testing.T) {
	server, captured, _ := newServer(t, 200, `{"data":[{"id":"claude-sonnet-5-5","display_name":"Claude Sonnet 5.5","max_input_tokens":1000000},{"id":"claude-haiku-5-5","display_name":"Claude Haiku 5.5"}],"has_more":false}`)
	models, err := New(Options{BaseURL: server.URL, APIKey: "k"}).ListModels(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if captured.URL.RequestURI() != "/models?limit=1000" || len(models) != 2 || models[0].ID != "claude-haiku-5-5" ||
		models[1].Name != "Claude Sonnet 5.5" || models[1].ContextWindow != 1000000 {
		t.Errorf("models = %+v (uri %s)", models, captured.URL.RequestURI())
	}
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
