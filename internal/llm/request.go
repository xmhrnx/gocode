package llm

import "encoding/json"

// Request is one model call.
type Request struct {
	Model    string
	System   string
	Messages []Message
	Tools    []ToolSpec
	// MaxTokens caps the reply length; 0 lets the adapter pick a default.
	MaxTokens int
	// Temperature is nil to use the provider's default.
	Temperature *float64
	// ReasoningEffort is "", "low", "medium" or "high". Empty leaves
	// reasoning at the model's default.
	ReasoningEffort string
}

// ToolSpec describes a tool the model may call.
type ToolSpec struct {
	Name        string
	Description string
	// Parameters is the JSON Schema of the tool's arguments object.
	Parameters json.RawMessage
}
