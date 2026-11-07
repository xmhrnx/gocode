package gemini

import (
	"encoding/json"
	"strings"
)

// wireContent is one conversation turn, sent and received.
type wireContent struct {
	Role  string     `json:"role,omitempty"`
	Parts []wirePart `json:"parts"`
}

// wirePart is a Part; exactly one payload is set. Text is a pointer so a
// part holding only a thought signature still carries "text":"".
type wirePart struct {
	Text             *string           `json:"text,omitempty"`
	Thought          bool              `json:"thought,omitempty"`
	ThoughtSignature string            `json:"thoughtSignature,omitempty"`
	InlineData       *inlineData       `json:"inlineData,omitempty"`
	FunctionCall     *functionCall     `json:"functionCall,omitempty"`
	FunctionResponse *functionResponse `json:"functionResponse,omitempty"`
	// raw is the part as received, kept for parts of the API's own tools
	// (a Google Search call and its response), which go back unchanged.
	raw json.RawMessage
}

// serverPart has the fields of the parts the API's own tools produce.
type serverPart struct {
	ToolCall *struct {
		Args struct {
			Query   string   `json:"query"`
			Queries []string `json:"queries"`
		} `json:"args"`
	} `json:"toolCall"`
	ToolResponse        json.RawMessage `json:"toolResponse"`
	ExecutableCode      json.RawMessage `json:"executableCode"`
	CodeExecutionResult json.RawMessage `json:"codeExecutionResult"`
}

type inlineData struct {
	MimeType string `json:"mimeType"`
	// Data is base64; encoding/json encodes []byte that way.
	Data []byte `json:"data"`
}

type functionCall struct {
	ID   string          `json:"id,omitempty"`
	Name string          `json:"name"`
	Args json.RawMessage `json:"args,omitempty"`
}

type functionResponse struct {
	ID       string         `json:"id,omitempty"`
	Name     string         `json:"name"`
	Response map[string]any `json:"response"`
}

// UnmarshalJSON keeps the part's JSON alongside its fields.
func (p *wirePart) UnmarshalJSON(data []byte) error {
	type plain wirePart
	if err := json.Unmarshal(data, (*plain)(p)); err != nil {
		return err
	}
	p.raw = append(json.RawMessage(nil), data...)
	return nil
}

// MarshalJSON sends a kept server part back exactly as received.
func (p wirePart) MarshalJSON() ([]byte, error) {
	if p.raw != nil {
		return p.raw, nil
	}
	type plain wirePart
	return json.Marshal(plain(p))
}

// serverDescription reports whether the part comes from the API's own
// tools and, for a search, what it searched for.
func (p wirePart) serverDescription() (string, bool) {
	var server serverPart
	if p.raw == nil || json.Unmarshal(p.raw, &server) != nil {
		return "", false
	}
	switch {
	case server.ToolCall != nil:
		queries := server.ToolCall.Args.Queries
		if server.ToolCall.Args.Query != "" {
			queries = append(queries, server.ToolCall.Args.Query)
		}
		if len(queries) == 0 {
			return "", true
		}
		return "Searched the web: " + strings.Join(queries, "; "), true
	case server.ToolResponse != nil || server.ExecutableCode != nil || server.CodeExecutionResult != nil:
		return "", true
	}
	return "", false
}

func textPtr(text string) *string { return &text }

// text returns the part's text, or "" when it has none.
func (p wirePart) text() string {
	if p.Text == nil {
		return ""
	}
	return *p.Text
}
