package mcp

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// callToolResult is what tools/call returns.
type callToolResult struct {
	Content           []contentItem   `json:"content"`
	StructuredContent json.RawMessage `json:"structuredContent,omitempty"`
	IsError           bool            `json:"isError"`
}

// contentItem is one block of a tool's output: text, image, audio, an
// embedded resource or a link to one.
type contentItem struct {
	Type     string            `json:"type"`
	Text     string            `json:"text"`
	Data     string            `json:"data"`
	MimeType string            `json:"mimeType"`
	URI      string            `json:"uri"`
	Resource *embeddedResource `json:"resource,omitempty"`
}

// embeddedResource is a resource's contents included in a result, as text
// or as a base64 blob.
type embeddedResource struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text"`
	Blob     string `json:"blob"`
}

// text renders the result for the model. Binary items become short
// placeholders, since the model cannot use raw base64. Structured content
// is shown only when no text carries the answer.
func (r callToolResult) text() string {
	parts := make([]string, 0, len(r.Content)+1)
	hasText := false
	for _, item := range r.Content {
		hasText = hasText || (item.Type == "text" && item.Text != "")
		if rendered := item.render(); rendered != "" {
			parts = append(parts, rendered)
		}
	}
	if !hasText && hasValue(r.StructuredContent) {
		parts = append(parts, string(r.StructuredContent))
	}
	return strings.Join(parts, "\n")
}

// render turns one content item into text.
func (item contentItem) render() string {
	switch item.Type {
	case "text":
		return item.Text
	case "image", "audio":
		return fmt.Sprintf("[%s: %s, %s]", item.Type, item.MimeType, base64Size(item.Data))
	case "resource":
		return item.renderResource()
	case "resource_link":
		return fmt.Sprintf("[resource link: %s]", item.URI)
	default:
		return fmt.Sprintf("[%s content]", item.Type)
	}
}

// renderResource shows an embedded resource's text, or describes a blob.
func (item contentItem) renderResource() string {
	res := item.Resource
	if res == nil {
		return "[resource]"
	}
	if res.Text != "" {
		return res.Text
	}
	return fmt.Sprintf("[resource: %s, %s, %s]", res.URI, res.MimeType, base64Size(res.Blob))
}

// hasValue reports whether raw holds a JSON value other than null.
func hasValue(raw json.RawMessage) bool {
	trimmed := strings.TrimSpace(string(raw))
	return trimmed != "" && trimmed != "null"
}

// base64Size describes how large the decoded form of data is.
func base64Size(data string) string {
	size := base64.StdEncoding.DecodedLen(len(data))
	switch {
	case size < 1<<10:
		return fmt.Sprintf("%d B", size)
	case size < 1<<20:
		return fmt.Sprintf("%d KB", (size+1<<9)>>10)
	default:
		return fmt.Sprintf("%.1f MB", float64(size)/(1<<20))
	}
}
