package gemini

import (
	"encoding/json"
	"strings"

	"gocode/internal/llm"
)

// generateRequest is the body of streamGenerateContent.
type generateRequest struct {
	Contents          []wireContent    `json:"contents"`
	SystemInstruction *wireContent     `json:"systemInstruction,omitempty"`
	Tools             []wireToolset    `json:"tools,omitempty"`
	ToolConfig        *toolConfig      `json:"toolConfig,omitempty"`
	GenerationConfig  generationConfig `json:"generationConfig"`
}

// wireToolset is one entry of tools: our functions or Google Search.
type wireToolset struct {
	FunctionDeclarations []functionDeclaration `json:"functionDeclarations,omitempty"`
	GoogleSearch         *struct{}             `json:"googleSearch,omitempty"`
}

// toolConfig lets Google Search run beside our functions: the API then
// returns the search as parts that must be sent back, and only checks
// function calls in VALIDATED mode.
type toolConfig struct {
	IncludeServerSideToolInvocations bool                   `json:"includeServerSideToolInvocations"`
	FunctionCallingConfig            *functionCallingConfig `json:"functionCallingConfig,omitempty"`
}

type functionCallingConfig struct {
	Mode string `json:"mode"`
}

// providerName marks the parts this adapter makes, so it sends back only
// its own.
const providerName = "gemini"

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
	specs := req.Tools
	search := req.WebSearch && canSearch(req.Model, len(specs) > 0)
	if search {
		specs = llm.WithoutTool(specs, llm.WebSearchTool)
	}
	if len(specs) > 0 {
		body.Tools = []wireToolset{{FunctionDeclarations: convertTools(specs)}}
	}
	if search {
		body.Tools = append(body.Tools, wireToolset{GoogleSearch: &struct{}{}})
		if len(specs) > 0 {
			body.ToolConfig = &toolConfig{IncludeServerSideToolInvocations: true,
				FunctionCallingConfig: &functionCallingConfig{Mode: "VALIDATED"}}
		}
	}
	return body
}

// canSearch reports whether Google Search can go in the request: on its
// own with any model, beside functions only from Gemini 3 on.
func canSearch(model string, withFunctions bool) bool {
	model = strings.TrimPrefix(model, "models/")
	return !withFunctions || strings.HasPrefix(model, "gemini-3")
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
