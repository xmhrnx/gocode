// Package agent runs the agent loop: it sends the conversation to the
// model, runs the tools the model calls, and repeats until the model
// answers without calling a tool.
package agent

import (
	"context"

	"gocode/internal/llm"
	"gocode/internal/tools"
)

// EventKind identifies what an Event reports.
type EventKind int

// Event is one thing that happened during a run.
type Event struct {
	Kind    EventKind
	Text    string
	Call    *llm.ToolCall
	Summary string
	Result  *tools.Result
	Message *llm.Message
	Usage   *llm.Usage
	// Parent is the ID of the task call whose sub-agent produced the
	// event; empty for the main agent's own events.
	Parent string
}

// Decision is the user's answer to a permission request.
type Decision int

// PermissionRequest asks the user to approve one tool call.
type PermissionRequest struct {
	Call    llm.ToolCall
	Access  tools.Access
	Summary string
	// Preview shows what the call would change (a diff), if known.
	Preview string
}

// PermissionAnswer is the user's decision, with optional feedback that is
// passed to the model when the call is denied.
type PermissionAnswer struct {
	Decision Decision
	Feedback string
}

// Asker shows a permission request and blocks until the user answers or
// ctx ends.
type Asker func(ctx context.Context, req PermissionRequest) PermissionAnswer

const (
	// EventText carries reply text in Text.
	EventText EventKind = iota
	// EventReasoning carries reasoning text in Text.
	EventReasoning
	// EventToolStarted reports a tool call about to run, in Call.
	EventToolStarted
	// EventToolFinished reports a finished tool call: Call and Result.
	EventToolFinished
	// EventMessage reports a message added to the conversation.
	EventMessage
	// EventUsage reports a model call's token usage.
	EventUsage
	// EventRetry reports a failed model call that will be retried, in Text.
	EventRetry
	// EventNotice reports something the user should know, in Text.
	EventNotice
	// EventCompacted reports that the conversation sent to the model was
	// replaced by the summary in Text.
	EventCompacted
)

const (
	DecisionAllow Decision = iota
	DecisionAllowAlways
	DecisionDeny
)
