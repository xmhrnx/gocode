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
	// WebSearch asks the provider to search the web itself, where it can.
	// An adapter that does leaves out a tool of ours named WebSearchTool.
	WebSearch bool
}

// WebSearchTool is the name of GoCode's own search tool, which a
// provider's built-in search replaces.
const WebSearchTool = "web_search"

// StopPaused is the stop reason of a reply the provider paused in the
// middle of its own tool use; sending the conversation again resumes it.
const StopPaused = "pause_turn"

// WithoutTool returns specs without the tool called name.
func WithoutTool(specs []ToolSpec, name string) []ToolSpec {
	out := make([]ToolSpec, 0, len(specs))
	for _, spec := range specs {
		if spec.Name != name {
			out = append(out, spec)
		}
	}
	return out
}

// ToolSpec describes a tool the model may call.
type ToolSpec struct {
	Name        string
	Description string
	// Parameters is the JSON Schema of the tool's arguments object.
	Parameters json.RawMessage
}
