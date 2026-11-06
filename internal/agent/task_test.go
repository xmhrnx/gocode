package agent

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"gocode/internal/llm"
	"gocode/internal/tools"
)

func taskCallEvent(prompt string) llm.Event {
	args, _ := json.Marshal(taskArgs{Description: "look around", Prompt: prompt})
	return llm.Event{Kind: llm.EventToolCall, ToolCall: &llm.ToolCall{ID: "task1", Name: TaskToolName, Arguments: args}}
}

func TestTaskRunsASubAgentAndReturnsItsReport(t *testing.T) {
	client := &scriptedClient{replies: [][]llm.Event{
		{taskCallEvent("find the config")},                        // main agent
		{toolCallEvent("c1", "inner")},                            // sub-agent uses a tool
		{{Kind: llm.EventText, Text: "Found it in config.go:12"}}, // sub-agent reports
		{{Kind: llm.EventText, Text: "Done."}},                    // main agent
	}}
	cfg := newTestConfig(client, tools.AccessRead, nil)
	cfg.Subagents, cfg.System, cfg.SubagentPrompt = true, "main prompt", "sub-agent rules"
	var tagged []Event
	added, err := Run(context.Background(), cfg, nil, func(ev Event) {
		if ev.Parent != "" {
			tagged = append(tagged, ev)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	result := added[1].Parts[0].ToolResult
	if result.Content != "Found it in config.go:12" || result.IsError {
		t.Fatalf("task result = %+v", result)
	}
	subRequest := client.requests[1]
	if !strings.Contains(subRequest.System, "sub-agent rules") || subRequest.Messages[0].Text() != "find the config" {
		t.Fatalf("sub-agent request: system %q, first message %q", subRequest.System, subRequest.Messages[0].Text())
	}
	for _, spec := range subRequest.Tools {
		if spec.Name == TaskToolName {
			t.Fatal("a sub-agent must not get the task tool")
		}
	}
	if len(tagged) == 0 || tagged[0].Parent != "task1" {
		t.Fatalf("sub-agent events not tagged with the task call: %+v", tagged)
	}
}
