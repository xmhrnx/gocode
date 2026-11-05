// Package llm defines the provider-neutral conversation model that every
// provider adapter translates to and from.
package llm

import "encoding/json"

// Role is who authored a message.
type Role string

// PartKind identifies what a Part carries.
type PartKind string

// Message is one turn of the conversation. Tool results travel in user
// messages, the way Anthropic models them; adapters for other APIs split
// them out as needed.
type Message struct {
	Role  Role   `json:"role"`
	Parts []Part `json:"parts"`
}

// Part is one piece of a message. Exactly one of the payload fields is set,
// matching Kind.
type Part struct {
	Kind       PartKind    `json:"kind"`
	Text       string      `json:"text,omitempty"`
	Image      *Image      `json:"image,omitempty"`
	ToolCall   *ToolCall   `json:"tool_call,omitempty"`
	ToolResult *ToolResult `json:"tool_result,omitempty"`
	// Signature authenticates a reasoning part; some providers require it
	// to be sent back unchanged on the next request.
	Signature string `json:"signature,omitempty"`
	// Redacted marks an encrypted reasoning part whose opaque payload is in
	// Signature; it must be sent back but can't be shown.
	Redacted bool `json:"redacted,omitempty"`
}

// Image is an inline image attachment.
type Image struct {
	MediaType string `json:"media_type"`
	Data      []byte `json:"data"`
}

// ToolCall is the model asking to run a tool.
type ToolCall struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
	// Signature is an opaque provider token tied to the call (Gemini's
	// thought signature) that must be echoed back.
	Signature string `json:"signature,omitempty"`
}

// ToolResult is the outcome of a tool call, sent back to the model.
type ToolResult struct {
	CallID  string `json:"call_id"`
	Name    string `json:"name"`
	Content string `json:"content"`
	IsError bool   `json:"is_error,omitempty"`
}

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

const (
	PartText       PartKind = "text"
	PartImage      PartKind = "image"
	PartReasoning  PartKind = "reasoning"
	PartToolCall   PartKind = "tool_call"
	PartToolResult PartKind = "tool_result"
)
