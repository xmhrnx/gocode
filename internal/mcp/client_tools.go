package mcp

import (
	"context"
	"encoding/json"
	"fmt"
)

// toolInfo describes one tool a server offers.
type toolInfo struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	InputSchema json.RawMessage  `json:"inputSchema"`
	Annotations *toolAnnotations `json:"annotations,omitempty"`
}

// toolAnnotations are the server's hints about a tool's behaviour.
type toolAnnotations struct {
	ReadOnlyHint bool `json:"readOnlyHint"`
}

// listToolsParams asks for one page of tools.
type listToolsParams struct {
	Cursor string `json:"cursor,omitempty"`
}

// listToolsResult is one page of tools.
type listToolsResult struct {
	Tools      []toolInfo `json:"tools"`
	NextCursor string     `json:"nextCursor"`
}

// callToolParams invokes a tool.
type callToolParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// listTools fetches every page of the server's tools.
func (c *client) listTools(ctx context.Context) ([]toolInfo, error) {
	var all []toolInfo
	seen := make(map[string]bool)
	cursor := ""
	for {
		var page listToolsResult
		if err := c.request(ctx, "tools/list", listToolsParams{Cursor: cursor}, &page); err != nil {
			return nil, err
		}
		all = append(all, page.Tools...)
		if page.NextCursor == "" {
			return all, nil
		}
		// A server repeating a cursor would otherwise loop forever.
		if seen[page.NextCursor] {
			return nil, fmt.Errorf("tools/list: server repeated cursor %q", page.NextCursor)
		}
		seen[page.NextCursor] = true
		cursor = page.NextCursor
	}
}

// callTool runs a tool with args, a JSON object.
func (c *client) callTool(ctx context.Context, name string, args json.RawMessage) (callToolResult, error) {
	var result callToolResult
	err := c.request(ctx, "tools/call", callToolParams{Name: name, Arguments: args}, &result)
	return result, err
}
