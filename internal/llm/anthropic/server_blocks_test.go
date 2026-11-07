package anthropic

import (
	"encoding/json"
	"strings"
	"testing"

	"gocode/internal/llm"
)

var searchStream = sseEvent("message_start", `{"type":"message_start","message":{"usage":{"input_tokens":5}}}`) +
	sseEvent("content_block_start", `{"type":"content_block_start","index":0,"content_block":{"type":"server_tool_use","id":"srvtoolu_1","name":"web_search","input":{}}}`) +
	sseEvent("content_block_delta", `{"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"query\": \"go 1.26"}}`) +
	sseEvent("content_block_delta", `{"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":" release\"}"}}`) +
	sseEvent("content_block_stop", `{"type":"content_block_stop","index":0}`) +
	sseEvent("content_block_start", `{"type":"content_block_start","index":1,"content_block":{"type":"web_search_tool_result","tool_use_id":"srvtoolu_1","content":[{"type":"web_search_result","url":"https://go.dev","title":"Go","encrypted_content":"abc"}]}}`) +
	sseEvent("content_block_stop", `{"type":"content_block_stop","index":1}`) +
	sseEvent("message_delta", `{"type":"message_delta","delta":{"stop_reason":"pause_turn"},"usage":{"output_tokens":9}}`) +
	sseEvent("message_stop", `{"type":"message_stop"}`)

func TestServerSearchBlocksBecomeProviderParts(t *testing.T) {
	server, _, body := newServer(t, 200, searchStream)
	req := llm.Request{Model: "claude-opus-5-5", WebSearch: true,
		Tools: []llm.ToolSpec{{Name: "web_search"}, {Name: "read_file"}}}
	events, err := collect(t, New(Options{BaseURL: server.URL, APIKey: "k"}), req)
	if err != nil {
		t.Fatal(err)
	}

	tools, _ := json.Marshal(body["tools"])
	if strings.Count(string(tools), `"web_search"`) != 1 || !strings.Contains(string(tools), "web_search_20260209") {
		t.Fatalf("tools sent: %s", tools)
	}

	var parts []llm.Part
	stop := ""
	for _, e := range events {
		switch e.Kind {
		case llm.EventPart:
			parts = append(parts, *e.Part)
		case llm.EventStop:
			stop = e.StopReason
		}
	}
	if len(parts) != 2 || stop != llm.StopPaused {
		t.Fatalf("parts %d, stop %q", len(parts), stop)
	}
	if parts[0].Text != "Searched the web: go 1.26 release" || parts[1].Text != "" {
		t.Fatalf("descriptions %q %q", parts[0].Text, parts[1].Text)
	}
	if !strings.Contains(string(parts[0].Raw), `"query":"go 1.26 release"`) {
		t.Fatalf("input not put into the block: %s", parts[0].Raw)
	}

	// Sent back unchanged in the next request, and only to this provider.
	history := []llm.Message{{Role: llm.RoleAssistant, Parts: append(parts,
		llm.Part{Kind: llm.PartProvider, Provider: "gemini", Raw: json.RawMessage(`{"x":1}`)})}}
	sent, _ := json.Marshal(convertMessages(history))
	if !strings.Contains(string(sent), `"encrypted_content":"abc"`) || strings.Contains(string(sent), `"x":1`) {
		t.Fatalf("replayed: %s", sent)
	}
}

func TestWebSearchToolVersion(t *testing.T) {
	if got := webSearchTool("claude-haiku-4-5")["type"]; got != "web_search_20250305" {
		t.Fatalf("haiku 4.5 gets %s", got)
	}
	if got := webSearchTool("claude-sonnet-5-5")["type"]; got != "web_search_20260209" {
		t.Fatalf("sonnet 5.5 gets %s", got)
	}
}
