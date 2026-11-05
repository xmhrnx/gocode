package agent

import (
	"context"
	"strings"
	"testing"

	"gocode/internal/llm"
	"gocode/internal/settings"
	"gocode/internal/tools"
)

// overflowingClient rejects the first request as too long, then answers.
type overflowingClient struct {
	calls    int
	requests []llm.Request
}

func (c *overflowingClient) Stream(_ context.Context, req llm.Request, emit func(llm.Event)) error {
	c.calls++
	c.requests = append(c.requests, req)
	switch c.calls {
	case 1:
		return &llm.APIError{StatusCode: 400, Message: "This model's maximum context length is 262144 tokens"}
	case 2:
		emit(llm.Event{Kind: llm.EventText, Text: "summary of the work"})
	default:
		emit(llm.Event{Kind: llm.EventText, Text: "continuing"})
	}
	return nil
}

func (c *overflowingClient) ListModels(context.Context) ([]llm.ModelInfo, error) { return nil, nil }

func TestRunCompactsAndRetriesWhenTheContextOverflows(t *testing.T) {
	client := &overflowingClient{}
	cfg := Config{Client: client, ContextWindow: 1000, CompactAtPercent: 80, CompactPrompt: "summarize"}
	history := []llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart("big task")}}}
	var compacted string
	added, err := Run(context.Background(), cfg, history, func(ev Event) {
		if ev.Kind == EventCompacted {
			compacted = ev.Text
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if compacted != "summary of the work" || added[0].Text() != "continuing" {
		t.Fatalf("compacted=%q added=%v", compacted, added)
	}
	retried := client.requests[2].Messages
	if len(retried) != 1 || !strings.Contains(retried[0].Text(), "summary of the work") {
		t.Fatalf("retry was not sent with the summary: %+v", retried)
	}
}

func TestNeedsCompaction(t *testing.T) {
	cfg := Config{ContextWindow: 1000, CompactAtPercent: 80}
	if cfg.needsCompaction(700) || !cfg.needsCompaction(800) {
		t.Fatal("threshold misplaced")
	}
	if (Config{ContextWindow: 1000}).needsCompaction(999) {
		t.Fatal("compaction should be off at 0 percent")
	}
}

func TestRepeatedCallsAskEvenWhenAllowed(t *testing.T) {
	var replies [][]llm.Event
	for i := 0; i < repeatLimit; i++ {
		replies = append(replies, []llm.Event{toolCallEvent("c", "same")})
	}
	replies = append(replies, []llm.Event{{Kind: llm.EventText, Text: "done"}})
	asked := 0
	ask := func(context.Context, PermissionRequest) PermissionAnswer {
		asked++
		return PermissionAnswer{Decision: DecisionAllow}
	}
	cfg := newTestConfig(&scriptedClient{replies: replies}, tools.AccessRead, ask)
	cfg.Permissions = NewPermissions(ModeFullAccess, settings.Default().Agent)
	if _, err := Run(context.Background(), cfg, nil, func(Event) {}); err != nil {
		t.Fatal(err)
	}
	if asked != 1 {
		t.Fatalf("asked %d times, want once on the third repeat", asked)
	}
}
