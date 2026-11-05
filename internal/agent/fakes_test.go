package agent

import (
	"context"
	"encoding/json"

	"gocode/internal/llm"
	"gocode/internal/tools"
)

// scriptedClient replies with one scripted event list per call.
type scriptedClient struct {
	replies  [][]llm.Event
	requests []llm.Request
}

// echoTool returns its "text" argument.
type echoTool struct{ access tools.Access }

func (c *scriptedClient) Stream(ctx context.Context, req llm.Request, emit func(llm.Event)) error {
	c.requests = append(c.requests, req)
	reply := c.replies[0]
	c.replies = c.replies[1:]
	for _, ev := range reply {
		emit(ev)
	}
	return nil
}

func (c *scriptedClient) ListModels(context.Context) ([]llm.ModelInfo, error) { return nil, nil }

func (echoTool) Spec() llm.ToolSpec {
	return llm.ToolSpec{Name: "echo", Parameters: json.RawMessage(`{"type":"object"}`)}
}
func (t echoTool) Access() tools.Access         { return t.access }
func (echoTool) Summary(json.RawMessage) string { return "echo" }
func (echoTool) Run(_ context.Context, _ tools.Env, args json.RawMessage) tools.Result {
	var in struct{ Text string }
	_ = json.Unmarshal(args, &in)
	return tools.Result{Output: in.Text}
}

func toolCallEvent(id, text string) llm.Event {
	args, _ := json.Marshal(map[string]string{"text": text})
	return llm.Event{Kind: llm.EventToolCall, ToolCall: &llm.ToolCall{ID: id, Name: "echo", Arguments: args}}
}
