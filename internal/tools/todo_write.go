package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gocode/internal/llm"
)

// TodoWrite records the agent's task list.
type TodoWrite struct{}

type todoWriteArgs struct {
	Todos []Todo `json:"todos"`
}

// todoMarks renders each status as a checkbox.
var todoMarks = map[string]string{"pending": "[ ]", "in_progress": "[>]", "completed": "[x]"}

// Spec describes todo_write to the model.
func (TodoWrite) Spec() llm.ToolSpec {
	return llm.ToolSpec{
		Name: "todo_write",
		Description: "Create or update your task list, shown to the user. Use it for tasks with three or more steps or when the user gives several tasks; " +
			"skip it for simple one-step requests. Each call replaces the whole list. " +
			"Keep exactly one item in_progress while working, mark items completed as soon as they are done, and add items you discover along the way. " +
			"Before your final reply, call it once more so the list shows what was done: every finished item completed, and items you won't do removed.",
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "todos": {
      "type": "array",
      "description": "The complete, updated task list.",
      "items": {
        "type": "object",
        "properties": {
          "content": {"type": "string", "description": "The task, as a short imperative sentence."},
          "status": {"type": "string", "enum": ["pending", "in_progress", "completed"], "description": "The task's state."}
        },
        "required": ["content", "status"],
        "additionalProperties": false
      }
    }
  },
  "required": ["todos"],
  "additionalProperties": false
}`),
	}
}

// Access reports that todo_write changes nothing in the workspace.
func (TodoWrite) Access() Access { return AccessRead }

// Summary counts the items.
func (TodoWrite) Summary(raw json.RawMessage) string {
	var args todoWriteArgs
	summaryArgs(raw, &args)
	return fmt.Sprintf("Update todo list (%d items)", len(args.Todos))
}

// Run validates the list and hands it to the UI.
func (TodoWrite) Run(ctx context.Context, env Env, raw json.RawMessage) Result {
	var args todoWriteArgs
	if err := decodeArgs(raw, &args); err != nil {
		return Errorf("%v", err)
	}
	if err := validateTodos(args.Todos); err != nil {
		return Errorf("%v", err)
	}
	if env.Todos != nil {
		env.Todos(append([]Todo(nil), args.Todos...))
	}
	return Result{Output: "Todo list updated.\n" + renderTodos(args.Todos)}
}

// validateTodos checks every item and that at most one is in progress.
func validateTodos(todos []Todo) error {
	inProgress := 0
	for i, todo := range todos {
		if strings.TrimSpace(todo.Content) == "" {
			return fmt.Errorf("todo %d has empty content", i+1)
		}
		if _, ok := todoMarks[todo.Status]; !ok {
			return fmt.Errorf("todo %d has status %q; use pending, in_progress or completed", i+1, todo.Status)
		}
		if todo.Status == "in_progress" {
			inProgress++
		}
	}
	if inProgress > 1 {
		return fmt.Errorf("%d todos are in_progress; keep exactly one in_progress at a time", inProgress)
	}
	return nil
}

// renderTodos lists the items with checkbox marks.
func renderTodos(todos []Todo) string {
	if len(todos) == 0 {
		return "(empty list)"
	}
	var out strings.Builder
	for _, todo := range todos {
		fmt.Fprintf(&out, "%s %s\n", todoMarks[todo.Status], todo.Content)
	}
	return strings.TrimSuffix(out.String(), "\n")
}
