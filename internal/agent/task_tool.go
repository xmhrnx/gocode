package agent

import (
	"context"
	"encoding/json"

	"gocode/internal/llm"
	"gocode/internal/tools"
)

// taskTool runs a sub-agent on a self-contained job and returns its
// report. The sub-agent works in a fresh context, so the files it reads
// and the commands it runs don't fill the main conversation.
type taskTool struct {
	parent Config
}

// taskArgs are the task tool's arguments.
type taskArgs struct {
	Description string `json:"description"`
	Prompt      string `json:"prompt"`
}

// TaskToolName is the task tool's name, for settings that switch it off.
const TaskToolName = "task"

// TaskSpec describes the task tool to the model.
func TaskSpec() llm.ToolSpec {
	return llm.ToolSpec{
		Name: TaskToolName,
		Description: "Start a sub-agent that does one self-contained task with the same tools and returns a short report. " +
			"It starts with an empty context: put everything it needs in the prompt (goal, relevant paths, constraints, what to report). " +
			"Use it for work that needs a lot of reading or searching but only a little of the result, such as finding where something is " +
			"implemented across a large codebase, reviewing many files, or investigating a failure, so your own context stays small. " +
			"Don't use it for simple lookups you can do with one or two tool calls. Its report is not shown to the user; summarize what matters.",
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "description": {"type": "string", "description": "What the sub-agent does, in 3-6 words, shown to the user."},
    "prompt": {"type": "string", "description": "Complete instructions for the sub-agent."}
  },
  "required": ["description", "prompt"],
  "additionalProperties": false
}`),
	}
}

// Spec implements tools.Tool.
func (taskTool) Spec() llm.ToolSpec { return TaskSpec() }

// Access is read: starting a sub-agent does nothing by itself, and every
// tool call it makes goes through the permission policy as usual.
func (taskTool) Access() tools.Access { return tools.AccessRead }

// Summary shows the task's short description.
func (taskTool) Summary(raw json.RawMessage) string {
	var args taskArgs
	_ = json.Unmarshal(raw, &args) // a summary of bad arguments is just "Task"
	if args.Description == "" {
		return "Task"
	}
	return "Task: " + args.Description
}

// TaskSummary describes a task call in one line, for the UI.
func TaskSummary(raw json.RawMessage) string { return taskTool{}.Summary(raw) }

// Run is never called: the agent loop runs task calls with runCall, which
// can report the sub-agent's progress.
func (taskTool) Run(context.Context, tools.Env, json.RawMessage) tools.Result {
	return tools.Errorf("the task tool runs only inside the agent loop")
}
