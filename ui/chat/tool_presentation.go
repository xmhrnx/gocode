package chat

import (
	"encoding/json"
	"strings"

	"gocode/internal/agent"
)

// toolIcons maps tool names to their icons; unknown tools get a wrench.
var toolIcons = map[string]string{
	"read_file": "file-text", "write_file": "file-plus", "edit_file": "file-pen-line",
	"list_dir": "list-tree", "glob": "folder-open", "grep": "text-search",
	"bash": "terminal", "web_fetch": "globe", "todo_write": "list-todo",
	"task": "bot", "skill": "sparkles",
}

func toolIcon(name string) string {
	if icon, ok := toolIcons[name]; ok {
		return icon
	}
	return "wrench"
}

// toolSummary describes a call in one line, using the tool's own summary
// when the tool is known.
func (c *ChatTab) toolSummary(name string, args json.RawMessage) string {
	if name == agent.TaskToolName {
		return agent.TaskSummary(args) // added by the agent loop, not in the registry
	}
	if tool, ok := c.registry.Lookup(name); ok {
		return tool.Summary(args)
	}
	return name
}

// maxShownLines bounds tool output in the transcript; the full output is
// still what the model saw.
const maxShownLines = 400

// clipLines returns at most limit lines of text and how many were left out.
func clipLines(text string, limit int) (string, int) {
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	if len(lines) <= limit {
		return strings.Join(lines, "\n"), 0
	}
	return strings.Join(lines[:limit], "\n"), len(lines) - limit
}
