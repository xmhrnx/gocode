package gemini

import "encoding/json"

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

func textPtr(text string) *string { return &text }

// text returns the part's text, or "" when it has none.
func (p wirePart) text() string {
	if p.Text == nil {
		return ""
	}
	return *p.Text
}
