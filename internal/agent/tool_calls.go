package agent

import (
	"context"
	"fmt"

	"gocode/internal/llm"
	"gocode/internal/settings"
	"gocode/internal/tools"
)

const cancelledOutput = "The user cancelled this tool call."

// runToolCalls runs each call in order and returns a message with all the
// results. When ctx ends, the calls not yet run get a cancelled result.
func runToolCalls(ctx context.Context, cfg Config, calls []llm.ToolCall, guard *loopGuard, emit func(Event)) (llm.Message, error) {
	msg := llm.Message{Role: llm.RoleUser}
	for i, call := range calls {
		if ctx.Err() != nil {
			msg.Parts = append(msg.Parts, cancelledResults(calls[i:]).Parts...)
			return msg, ctx.Err()
		}
		result := runToolCall(ctx, cfg, call, guard.observe(call), emit)
		msg.Parts = append(msg.Parts, llm.ToolResultPart(llm.ToolResult{
			CallID: call.ID, Name: call.Name, Content: result.Output, IsError: result.IsError,
		}))
	}
	return msg, ctx.Err()
}

func runToolCall(ctx context.Context, cfg Config, call llm.ToolCall, repeats int, emit func(Event)) tools.Result {
	tool, ok := cfg.Tools.Lookup(call.Name)
	if !ok || (cfg.ToolEnabled != nil && !cfg.ToolEnabled(tool)) {
		result := tools.Errorf("Unknown tool %q.", call.Name)
		emit(Event{Kind: EventToolFinished, Call: &call, Result: &result})
		return result
	}
	summary := tool.Summary(call.Arguments)
	emit(Event{Kind: EventToolStarted, Call: &call, Summary: summary})
	result, approved := authorize(ctx, cfg, tool, call, summary, repeats)
	switch runner, isRunner := tool.(callRunner); {
	case approved && isRunner:
		result = runner.runCall(ctx, call, emit)
	case approved:
		result = tool.Run(ctx, cfg.Env, call.Arguments)
	}
	emit(Event{Kind: EventToolFinished, Call: &call, Summary: summary, Result: &result})
	return result
}

// authorize applies the permission policy, asking the user when needed.
// It returns the result to report when the call must not run.
func authorize(ctx context.Context, cfg Config, tool tools.Tool, call llm.ToolCall, summary string, repeats int) (tools.Result, bool) {
	policy := cfg.Permissions.Policy(tool)
	if repeats >= repeatLimit && policy == settings.PolicyAllow {
		policy = settings.PolicyAsk // a call stuck in a loop is worth a look, even when allowed
		summary = fmt.Sprintf("Repeated %d times in a row: %s", repeats, summary)
	}
	switch policy {
	case settings.PolicyAllow:
		return tools.Result{}, true
	case settings.PolicyDeny:
		return tools.Errorf("This action is not allowed in the current permission mode (%s). Do not try to achieve it another way.", cfg.Permissions.Mode()), false
	}
	answer := cfg.Ask(ctx, PermissionRequest{
		Call: call, Access: tool.Access(), Summary: summary, Preview: preview(tool, cfg.Env, call),
	})
	switch {
	case ctx.Err() != nil:
		return tools.Result{Output: cancelledOutput, IsError: true}, false
	case answer.Decision == DecisionAllowAlways:
		cfg.Permissions.AllowAlways(call.Name)
		return tools.Result{}, true
	case answer.Decision == DecisionAllow:
		return tools.Result{}, true
	}
	output := "The user denied this tool call."
	if answer.Feedback != "" {
		output += " Their feedback: " + answer.Feedback
	}
	return tools.Result{Output: output, IsError: true}, false
}

func preview(tool tools.Tool, env tools.Env, call llm.ToolCall) string {
	previewer, ok := tool.(tools.Previewer)
	if !ok {
		return ""
	}
	text, err := previewer.Preview(env, call.Arguments)
	if err != nil {
		return "" // the call itself will report the problem
	}
	return text
}

func cancelledResults(calls []llm.ToolCall) llm.Message {
	msg := llm.Message{Role: llm.RoleUser}
	for _, call := range calls {
		msg.Parts = append(msg.Parts, llm.ToolResultPart(llm.ToolResult{
			CallID: call.ID, Name: call.Name, Content: cancelledOutput, IsError: true,
		}))
	}
	return msg
}
