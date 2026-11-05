package mcp

import (
	"bytes"
	"context"
	"encoding/json"

	"gocode/internal/llm"
	"gocode/internal/tools"
)

// toolCaller runs a tool on an MCP server.
type toolCaller interface {
	callTool(ctx context.Context, name string, args json.RawMessage) (callToolResult, error)
}

// serverTool offers one MCP server tool to the agent as a tools.Tool.
type serverTool struct {
	caller toolCaller
	server string
	info   toolInfo
	name   string
}

const (
	// defaultMaxOutputBytes bounds output when Env.MaxOutputBytes is 0.
	defaultMaxOutputBytes = 30000
	// summaryArgsLimit bounds the arguments shown in a call summary.
	summaryArgsLimit = 80
)

// emptyObjectSchema is used for tools that declare no input schema, since
// providers require an object schema for every tool.
var emptyObjectSchema = json.RawMessage(`{"type":"object","properties":{}}`)

// Spec describes the tool with the server's description and schema.
func (t *serverTool) Spec() llm.ToolSpec {
	schema := t.info.InputSchema
	if !hasValue(schema) || string(bytes.TrimSpace(schema)) == "{}" {
		schema = emptyObjectSchema
	}
	return llm.ToolSpec{Name: t.name, Description: t.info.Description, Parameters: schema}
}

// Access is execute, so the user is asked by default: an MCP tool can do
// anything. Tools the server marks read-only are trusted as reads.
func (t *serverTool) Access() tools.Access {
	if t.info.Annotations != nil && t.info.Annotations.ReadOnlyHint {
		return tools.AccessRead
	}
	return tools.AccessExecute
}

// Summary names the server and tool, followed by the arguments on one line.
func (t *serverTool) Summary(args json.RawMessage) string {
	summary := t.server + ": " + t.info.Name
	if rendered := compactArgs(args); rendered != "" {
		summary += " " + oneLine(rendered, summaryArgsLimit)
	}
	return summary
}

// Run calls the tool on its server.
func (t *serverTool) Run(ctx context.Context, env tools.Env, args json.RawMessage) tools.Result {
	if !hasValue(args) {
		args = json.RawMessage(`{}`)
	}
	result, err := t.caller.callTool(ctx, t.info.Name, args)
	if err != nil {
		return tools.Errorf("%v", err)
	}
	limit := env.MaxOutputBytes
	if limit <= 0 {
		limit = defaultMaxOutputBytes
	}
	output := result.text()
	if output == "" {
		output = "(no output)"
	}
	return tools.Result{Output: truncateMiddle(output, limit), IsError: result.IsError}
}

// compactArgs renders args as compact JSON, or "" when there are none.
func compactArgs(args json.RawMessage) string {
	if !hasValue(args) {
		return ""
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, args); err != nil {
		return string(args) // not valid JSON; show it as the model sent it
	}
	if compact.String() == "{}" {
		return ""
	}
	return compact.String()
}
