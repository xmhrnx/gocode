package gemini

import (
	"encoding/json"

	"gocode/internal/llm"
)

// generateRequest is the body of streamGenerateContent.
type generateRequest struct {
	Contents          []wireContent    `json:"contents"`
	SystemInstruction *wireContent     `json:"systemInstruction,omitempty"`
	Tools             []wireToolset    `json:"tools,omitempty"`
	GenerationConfig  generationConfig `json:"generationConfig"`
}

type wireToolset struct {
	FunctionDeclarations []functionDeclaration `json:"functionDeclarations"`
}

// functionDeclaration uses parametersJsonSchema rather than parameters,
// which only accepts an OpenAPI subset of JSON Schema.
type functionDeclaration struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parametersJsonSchema,omitempty"`
}

type generationConfig struct {
	MaxOutputTokens int            `json:"maxOutputTokens,omitempty"`
	Temperature     *float64       `json:"temperature,omitempty"`
	ThinkingConfig  thinkingConfig `json:"thinkingConfig"`
}

type thinkingConfig struct {
	IncludeThoughts bool   `json:"includeThoughts"`
	ThinkingLevel   string `json:"thinkingLevel,omitempty"`
}

// thinkingLevels maps llm reasoning efforts to Gemini thinking levels.
var thinkingLevels = map[string]string{"low": "LOW", "medium": "MEDIUM", "high": "HIGH"}

func buildRequest(req llm.Request) generateRequest {
	body := generateRequest{
		Contents: convertContents(req.Messages),
		GenerationConfig: generationConfig{
			MaxOutputTokens: req.MaxTokens,
			Temperature:     req.Temperature,
			ThinkingConfig: thinkingConfig{
				IncludeThoughts: true,
				ThinkingLevel:   thinkingLevels[req.ReasoningEffort],
			},
		},
	}
	if req.System != "" {
		body.SystemInstruction = &wireContent{Parts: []wirePart{{Text: textPtr(req.System)}}}
	}
	if len(req.Tools) > 0 {
		body.Tools = []wireToolset{{FunctionDeclarations: convertTools(req.Tools)}}
	}
	return body
}

func convertTools(specs []llm.ToolSpec) []functionDeclaration {
	declarations := make([]functionDeclaration, 0, len(specs))
	for _, spec := range specs {
		declarations = append(declarations, functionDeclaration{
			Name: spec.Name, Description: spec.Description, Parameters: spec.Parameters,
		})
	}
	return declarations
}
