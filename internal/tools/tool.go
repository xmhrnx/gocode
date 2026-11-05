// Package tools holds the tools the agent can call: reading and editing
// files, searching, running commands and fetching web pages.
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"gocode/internal/llm"
)

// Access classifies what a tool can do, which decides whether it needs the
// user's permission.
type Access string

// Tool is one capability offered to the model.
type Tool interface {
	Spec() llm.ToolSpec
	Access() Access
	// Summary describes a call in one line for the UI and permission
	// prompts, e.g. "Edit internal/app.go" or "Run go test ./...".
	Summary(args json.RawMessage) string
	Run(ctx context.Context, env Env, args json.RawMessage) Result
}

// Previewer is implemented by tools that can show what a call would change
// before it runs (a diff of an edit), for permission prompts.
type Previewer interface {
	Preview(env Env, args json.RawMessage) (string, error)
}

// Env is the context a tool runs in.
type Env struct {
	// WorkDir is the project root; relative paths resolve against it.
	WorkDir string
	// Shell is the program used to run commands, e.g. "/bin/bash".
	Shell string
	// CommandTimeoutSeconds bounds a command's run time.
	CommandTimeoutSeconds int
	// MaxOutputBytes bounds what a tool returns to the model.
	MaxOutputBytes int
	// Todos receives the agent's task list from the todo tool; may be nil.
	Todos func([]Todo)
}

// Result is what a tool returns.
type Result struct {
	// Output is sent to the model.
	Output string
	// IsError marks a failed call.
	IsError bool
	// Display, when set, is shown in the UI instead of Output (a diff).
	Display string
}

// Todo is one item of the agent's task list.
type Todo struct {
	Content string `json:"content"`
	Status  string `json:"status"` // pending, in_progress or completed
}

const (
	// AccessRead only reads the workspace.
	AccessRead Access = "read"
	// AccessWrite changes files.
	AccessWrite Access = "write"
	// AccessExecute runs programs.
	AccessExecute Access = "execute"
	// AccessNetwork reaches the internet.
	AccessNetwork Access = "network"
)

// Errorf returns a failed Result.
func Errorf(format string, args ...any) Result {
	return Result{Output: fmt.Sprintf(format, args...), IsError: true}
}
