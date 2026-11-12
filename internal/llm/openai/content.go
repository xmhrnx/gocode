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
	// Images are pictures the model made (OpenRouter and others).
	Images []imageItem `json:"images"`
}

// imageItem is an image in a reply: a data: URL, or a link to it.
type imageItem struct {
	ImageURL struct {
		URL string `json:"url"`
	} `json:"image_url"`
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
	imageItem
}

// contentPiece is a run of reply text or reasoning, or an image.
type contentPiece struct {
	text      string
	reasoning bool
	imageURL  string
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
		case "image_url":
			pieces = append(pieces, contentPiece{imageURL: item.ImageURL.URL})
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
