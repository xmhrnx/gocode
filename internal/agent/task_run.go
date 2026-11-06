package agent

import (
	"context"
	"encoding/json"
	"strings"

	"gocode/internal/llm"
	"gocode/internal/tools"
)

// callRunner is a tool that runs with its call and the event stream: the
// task tool, whose sub-agent reports progress as it works.
type callRunner interface {
	runCall(ctx context.Context, call llm.ToolCall, emit func(Event)) tools.Result
}

// runCall runs the sub-agent and returns its final message as the result.
// Its events go to emit marked with the call's ID, so the UI can show its
// progress inside the task's card.
func (t taskTool) runCall(ctx context.Context, call llm.ToolCall, emit func(Event)) tools.Result {
	var args taskArgs
	if err := json.Unmarshal(call.Arguments, &args); err != nil || strings.TrimSpace(args.Prompt) == "" {
		return tools.Errorf("task needs a prompt with complete instructions for the sub-agent")
	}
	child := t.parent
	child.Subagents = false // one level only: a sub-agent can't start sub-agents
	child.Tools = t.parent.Tools.Without(TaskToolName)
	child.System = t.parent.System + "\n" + t.parent.SubagentPrompt
	child.Inbox = nil // the user's messages are for the main agent
	history := []llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart(args.Prompt)}}}
	added, err := Run(ctx, child, history, func(ev Event) {
		if ev.Parent == "" {
			ev.Parent = call.ID
		}
		emit(ev)
	})
	if err != nil {
		return tools.Errorf("The sub-agent stopped: %v. Its last words: %s", err, lastText(added))
	}
	return tools.Result{Output: report(added)}
}

// report is the sub-agent's final reply, which the prompt asks to be its
// report.
func report(added []llm.Message) string {
	if text := lastText(added); text != "" {
		return text
	}
	return "The sub-agent finished without a report."
}

func lastText(messages []llm.Message) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if text := strings.TrimSpace(messages[i].Text()); messages[i].Role == llm.RoleAssistant && text != "" {
			return text
		}
	}
	return ""
}
