package tools

import (
	"strings"
	"testing"
)

func TestTodoWrite(t *testing.T) {
	var received []Todo
	env := Env{Todos: func(todos []Todo) { received = todos }}
	todos := []Todo{{"Write code", "completed"}, {"Test it", "in_progress"}, {"Ship", "pending"}}
	res := runTool(t, TodoWrite{}, env, todoWriteArgs{Todos: todos})
	want := "Todo list updated.\n[x] Write code\n[>] Test it\n[ ] Ship"
	if res.IsError || res.Output != want {
		t.Errorf("got %q", res.Output)
	}
	if len(received) != 3 || received[1].Status != "in_progress" {
		t.Errorf("Env.Todos got %+v", received)
	}
}

func TestTodoWriteValidates(t *testing.T) {
	tests := []struct {
		todos []Todo
		want  string
	}{
		{[]Todo{{"a", "done"}}, "status \"done\""},
		{[]Todo{{" ", "pending"}}, "empty content"},
		{[]Todo{{"a", "in_progress"}, {"b", "in_progress"}}, "exactly one"},
	}
	for _, tt := range tests {
		res := runTool(t, TodoWrite{}, Env{}, todoWriteArgs{Todos: tt.todos})
		if !res.IsError || !strings.Contains(res.Output, tt.want) {
			t.Errorf("todos %+v: got %q", tt.todos, res.Output)
		}
	}
}

func TestDefaultRegistry(t *testing.T) {
	names := []string{"bash", "edit_file", "glob", "grep", "list_dir", "read_file", "todo_write", "web_fetch", "write_file"}
	specs := Default().Specs(nil)
	if len(specs) != len(names) {
		t.Fatalf("got %d tools", len(specs))
	}
	for i, spec := range specs {
		if spec.Name != names[i] {
			t.Errorf("tool %d is %q, want %q", i, spec.Name, names[i])
		}
	}
}
