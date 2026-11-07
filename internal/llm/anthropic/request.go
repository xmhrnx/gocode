package anthropic

import (
	"encoding/json"

	"gocode/internal/llm"
)

// messagesRequest is the body of POST /messages. There is deliberately no
// temperature: this API version rejects sampling parameters.
type messagesRequest struct {
	Model        string          `json:"model"`
	MaxTokens    int             `json:"max_tokens"`
	System       string          `json:"system,omitempty"`
	Messages     []wireMessage   `json:"messages"`
	Tools        []any           `json:"tools,omitempty"`
	Stream       bool            `json:"stream"`
	Thinking     *thinkingConfig `json:"thinking,omitempty"`
	OutputConfig *outputConfig   `json:"output_config,omitempty"`
	// CacheControl at the top level turns on automatic prompt caching,
	// which makes each step of a tool loop reuse the previous prefix.
	CacheControl *cacheControl `json:"cache_control,omitempty"`
}

type thinkingConfig struct {
	Type    string `json:"type"`
	Display string `json:"display"`
}

type outputConfig struct {
	Effort string `json:"effort"`
}

type cacheControl struct {
	Type string `json:"type"`
}

type wireTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"input_schema"`
}

// emptySchema stands in for tools that take no arguments; the API
// requires an input schema.
var emptySchema = json.RawMessage(`{"type":"object","properties":{}}`)

func buildRequest(req llm.Request) messagesRequest {
	body := messagesRequest{
		Model:        req.Model,
		MaxTokens:    req.MaxTokens,
		System:       req.System,
		Messages:     convertMessages(req.Messages),
		Stream:       true,
		CacheControl: &cacheControl{Type: "ephemeral"},
	}
	specs := req.Tools
	if req.WebSearch {
		specs = llm.WithoutTool(specs, llm.WebSearchTool)
	}
	body.Tools = convertTools(specs)
	if req.WebSearch {
		body.Tools = append(body.Tools, webSearchTool(req.Model))
	}
	if body.MaxTokens <= 0 {
		body.MaxTokens = defaultMaxTokens
	}
	if req.ReasoningEffort != "" {
		body.Thinking = &thinkingConfig{Type: "adaptive", Display: "summarized"}
		body.OutputConfig = &outputConfig{Effort: req.ReasoningEffort}
	}
	return body
}

func convertTools(specs []llm.ToolSpec) []any {
	tools := make([]any, 0, len(specs))
	for _, spec := range specs {
		schema := spec.Parameters
		if len(schema) == 0 {
			schema = emptySchema
		}
		tools = append(tools, wireTool{Name: spec.Name, Description: spec.Description, InputSchema: schema})
	}
	return tools
}
