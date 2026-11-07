package agent

import (
	"context"
	"encoding/json"
	"testing"

	"gocode/internal/llm"
	"gocode/internal/tools"
)

// A reply the provider paused during its own search is sent back as it is,
// so the provider carries on, and its search parts are kept.
func TestRunResumesAPausedReply(t *testing.T) {
	search := llm.Part{Kind: llm.PartProvider, Provider: "anthropic", Raw: json.RawMessage(`{"type":"server_tool_use"}`), Text: "Searched the web: go"}
	client := &scriptedClient{replies: [][]llm.Event{
		{{Kind: llm.EventPart, Part: &search}, {Kind: llm.EventStop, StopReason: llm.StopPaused}},
		{{Kind: llm.EventText, Text: "Found it."}, {Kind: llm.EventStop, StopReason: "end_turn"}},
	}}
	cfg := newTestConfig(client, tools.AccessRead, nil)
	cfg.WebSearch = true
	var parts int
	added, err := Run(context.Background(), cfg, nil, func(ev Event) {
		if ev.Kind == EventPart {
			parts++
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(client.requests) != 2 || len(added) != 2 || parts != 1 {
		t.Fatalf("requests %d, messages %d, parts %d", len(client.requests), len(added), parts)
	}
	if !client.requests[0].WebSearch || added[0].Parts[0].Kind != llm.PartProvider {
		t.Fatalf("first request %+v, first message %+v", client.requests[0], added[0])
	}
	if last := client.requests[1].Messages; len(last) != 1 || last[0].Role != llm.RoleAssistant {
		t.Fatalf("the paused reply wasn't sent back: %+v", last)
	}
}
