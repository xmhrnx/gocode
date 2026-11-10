package openai

import (
	"encoding/json"

	"gocode/internal/llm"
)

// chatRequest is the body of POST /chat/completions.
type chatRequest struct {
	Model               string         `json:"model"`
	Messages            []chatMessage  `json:"messages"`
	Tools               []any          `json:"tools,omitempty"`
	Stream              bool           `json:"stream"`
	StreamOptions       *streamOptions `json:"stream_options,omitempty"`
	MaxCompletionTokens int            `json:"max_completion_tokens,omitempty"`
	MaxTokens           int            `json:"max_tokens,omitempty"`
	Temperature         *float64       `json:"temperature,omitempty"`
	ReasoningEffort     string         `json:"reasoning_effort,omitempty"`
}

type streamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type chatTool struct {
	Type     string       `json:"type"`
	Function chatFunction `json:"function"`
}

type chatFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters"`
}

// emptySchema stands in for tools that take no arguments; the API
// requires a parameters object.
var emptySchema = json.RawMessage(`{"type":"object","properties":{}}`)

func (c *Client) buildRequest(req llm.Request) chatRequest {
	body := chatRequest{
		Model:         req.Model,
		Messages:      convertMessages(req.System, req.Messages, c.opts.EchoReasoningField, c.opts.FileInput),
		Stream:        true,
		StreamOptions: &streamOptions{IncludeUsage: true},
		Temperature:   req.Temperature,
	}
	specs := req.Tools
	search := req.WebSearch && c.opts.WebSearchTool != ""
	if search {
		specs = llm.WithoutTool(specs, llm.WebSearchTool)
	}
	body.Tools = convertTools(specs)
	if search {
		// The provider runs the search; nothing comes back to replay.
		body.Tools = append(body.Tools, map[string]string{"type": c.opts.WebSearchTool})
	}
	if req.MaxTokens > 0 {
		if c.opts.MaxTokensField == MaxTokens {
			body.MaxTokens = req.MaxTokens
		} else {
			body.MaxCompletionTokens = req.MaxTokens
		}
	}
	if c.opts.SendReasoningEffort {
		body.ReasoningEffort = req.ReasoningEffort
	}
	return body
}

func convertTools(specs []llm.ToolSpec) []any {
	tools := make([]any, 0, len(specs))
	for _, spec := range specs {
		params := spec.Parameters
		if len(params) == 0 {
			params = emptySchema
		}
		tools = append(tools, chatTool{
			Type:     "function",
			Function: chatFunction{Name: spec.Name, Description: spec.Description, Parameters: params},
		})
	}
	return tools
}
