package openai

import (
	"encoding/json"
	"testing"

	"gocode/internal/llm"
)

func toolLoopHistory() []llm.Message {
	call := llm.ToolCall{ID: "c1", Name: "read", Arguments: json.RawMessage(`{"path":"a"}`)}
	return []llm.Message{
		{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart("look"), llm.ImagePart("image/png", []byte{1, 2, 3})}},
		{Role: llm.RoleAssistant, Parts: []llm.Part{
			{Kind: llm.PartReasoning, Text: "need file"},
			llm.TextPart("Reading."),
			{Kind: llm.PartToolCall, ToolCall: &call},
		}},
		{Role: llm.RoleUser, Parts: []llm.Part{
			llm.TextPart("also this"),
			llm.ToolResultPart(llm.ToolResult{CallID: "c1", Name: "read", Content: "file body"}),
		}},
		{Role: llm.RoleAssistant}, // cancelled empty reply is dropped
	}
}

func TestBuildRequestConvertsToolLoop(t *testing.T) {
	temperature := 0.2
	client := New(Options{BaseURL: "http://x", MaxTokensField: MaxTokens, EchoReasoningField: "reasoning_content"})
	req := client.buildRequest(llm.Request{
		Model: "m", System: "be brief", Messages: toolLoopHistory(), MaxTokens: 50,
		Temperature: &temperature, ReasoningEffort: "high",
		Tools: []llm.ToolSpec{{Name: "read", Description: "Read a file"}},
	})
	got := mustJSON(t, req)
	want := `{"model":"m","messages":[` +
		`{"role":"system","content":"be brief"},` +
		`{"role":"user","content":[{"type":"text","text":"look"},{"type":"image_url","image_url":{"url":"data:image/png;base64,AQID"}}]},` +
		`{"content":"Reading.","reasoning_content":"need file","role":"assistant","tool_calls":[{"id":"c1","type":"function","function":{"name":"read","arguments":"{\"path\":\"a\"}"}}]},` +
		`{"role":"tool","content":"file body","tool_call_id":"c1"},` +
		`{"role":"user","content":"also this"}],` +
		`"tools":[{"type":"function","function":{"name":"read","description":"Read a file","parameters":{"type":"object","properties":{}}}}],` +
		`"stream":true,"stream_options":{"include_usage":true},"max_tokens":50,"temperature":0.2}`
	if got != want {
		t.Errorf("request JSON\n got: %s\nwant: %s", got, want)
	}
}

func TestBuildRequestDefaultsAndReasoningEffort(t *testing.T) {
	client := New(Options{BaseURL: "http://x", SendReasoningEffort: true})
	got := mustJSON(t, client.buildRequest(llm.Request{
		Model: "m", MaxTokens: 10, ReasoningEffort: "low",
		Messages: []llm.Message{{Role: llm.RoleAssistant, Parts: []llm.Part{
			{Kind: llm.PartToolCall, ToolCall: &llm.ToolCall{ID: "c", Name: "n"}},
		}}},
	}))
	want := `{"model":"m","messages":[{"role":"assistant","content":null,"tool_calls":[{"id":"c","type":"function","function":{"name":"n","arguments":"{}"}}]}],` +
		`"stream":true,"stream_options":{"include_usage":true},"max_completion_tokens":10,"reasoning_effort":"low"}`
	if got != want {
		t.Errorf("request JSON\n got: %s\nwant: %s", got, want)
	}
}

func TestAuthHeaderVariants(t *testing.T) {
	azure := New(Options{APIKey: "k", AuthHeader: "api-key"}).header()
	if azure.Get("api-key") != "k" || azure.Get("Authorization") != "" {
		t.Errorf("azure header = %v", azure)
	}
	if local := New(Options{}).header(); len(local) != 0 {
		t.Errorf("keyless header = %v", local)
	}
}
