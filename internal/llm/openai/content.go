package openai

import (
	"encoding/json"
	"strings"

	"gocode/internal/llm/jsonstr"
)

// chunkDelta is the incremental message in a streamed choice. Providers
// disagree on where reasoning goes, so every known field is accepted.
type chunkDelta struct {
	// Content is a string, or (Mistral) an array of text/thinking items.
	Content          json.RawMessage   `json:"content"`
	ReasoningContent json.RawMessage   `json:"reasoning_content"`
	Reasoning        json.RawMessage   `json:"reasoning"`
	ReasoningDetails []reasoningDetail `json:"reasoning_details"`
	ToolCalls        []toolCallDelta   `json:"tool_calls"`
}

// reasoningDetail is an OpenRouter reasoning_details entry.
type reasoningDetail struct {
	Text string `json:"text"`
}

// contentItem is one element of an array-valued content field.
type contentItem struct {
	Type     string          `json:"type"`
	Text     string          `json:"text"`
	Thinking json.RawMessage `json:"thinking"`
}

// contentPiece is a run of reply text or reasoning.
type contentPiece struct {
	text      string
	reasoning bool
}

// reasoningText returns the delta's reasoning. Routers such as OpenRouter
// repeat the same text in several fields, so only the first non-empty one
// counts.
func (d chunkDelta) reasoningText() string {
	if reasoning, _ := jsonstr.Decode(d.ReasoningContent); reasoning != "" {
		return reasoning
	}
	if reasoning, _ := jsonstr.Decode(d.Reasoning); reasoning != "" {
		return reasoning
	}
	var text strings.Builder
	for _, detail := range d.ReasoningDetails {
		text.WriteString(detail.Text)
	}
	return text.String()
}

// contentPieces decodes a string or array content field. Shapes it doesn't
// recognise carry nothing displayable, so they yield no pieces.
func contentPieces(raw json.RawMessage) []contentPiece {
	if text, ok := jsonstr.Decode(raw); ok {
		if text == "" {
			return nil
		}
		return []contentPiece{{text: text}}
	}
	var items []contentItem
	if json.Unmarshal(raw, &items) != nil {
		return nil
	}
	var pieces []contentPiece
	for _, item := range items {
		switch item.Type {
		case "text":
			pieces = append(pieces, contentPiece{text: item.Text})
		case "thinking":
			pieces = append(pieces, contentPiece{text: thinkingText(item.Thinking), reasoning: true})
		}
	}
	return pieces
}

// thinkingText reads a thinking field that is either a string or an array
// of {type:"text", text} items.
func thinkingText(raw json.RawMessage) string {
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return text
	}
	var items []contentItem
	if json.Unmarshal(raw, &items) != nil {
		return ""
	}
	var joined strings.Builder
	for _, item := range items {
		joined.WriteString(item.Text)
	}
	return joined.String()
}
