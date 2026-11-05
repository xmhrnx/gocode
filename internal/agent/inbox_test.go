package agent

import (
	"context"
	"testing"

	"gocode/internal/llm"
	"gocode/internal/tools"
)

func TestMessagesSentDuringARunReachTheModel(t *testing.T) {
	client := &scriptedClient{replies: [][]llm.Event{
		{toolCallEvent("1", "x")},
		{{Kind: llm.EventText, Text: "noted"}},
		{{Kind: llm.EventText, Text: "answered"}},
	}}
	cfg := newTestConfig(client, tools.AccessRead, nil)
	pending := [][]llm.Message{
		{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart("also do y")}}},
		{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart("and z")}}},
	}
	cfg.Inbox = func() []llm.Message {
		if len(pending) == 0 {
			return nil
		}
		next := pending[0]
		pending = pending[1:]
		return next
	}
	added, err := Run(context.Background(), cfg, nil, func(Event) {})
	if err != nil {
		t.Fatal(err)
	}
	results := added[1]
	if results.Parts[0].Kind != llm.PartToolResult || results.Text() != "also do y" {
		t.Fatalf("tool results message = %+v", results)
	}
	if added[3].Text() != "and z" || added[4].Text() != "answered" {
		t.Fatalf("a message sent as the model finished should start another turn: %+v", added)
	}
}
