package llm

import "strings"

// TextPart returns a text part.
func TextPart(text string) Part { return Part{Kind: PartText, Text: text} }

// ImagePart returns an image part.
func ImagePart(mediaType string, data []byte) Part {
	return Part{Kind: PartImage, Image: &Image{MediaType: mediaType, Data: data}}
}

// DocumentPart returns a document part.
func DocumentPart(doc Document) Part { return Part{Kind: PartDocument, Document: &doc} }

// DocumentText is how a document goes to a provider that can't take it:
// its text, or for audio and video a description, in the same <file>
// tags as an attached text file, with where the file is when known.
func DocumentText(doc Document) string {
	text := doc.Text
	switch {
	case doc.IsMedia():
		text += " This model can't receive the file's content itself."
	case strings.TrimSpace(text) == "":
		text = "(No text could be extracted from this document; it may be scanned images.)"
	}
	if doc.Path != "" {
		text += "\nThe file is at " + doc.Path
	}
	return "<file name=\"" + doc.Name + "\">\n" + text + "\n</file>"
}

// IsMedia reports whether the document is audio or video.
func (d Document) IsMedia() bool {
	return strings.HasPrefix(d.MediaType, "audio/") || strings.HasPrefix(d.MediaType, "video/")
}

// ToolResultPart returns a tool result part.
func ToolResultPart(result ToolResult) Part {
	return Part{Kind: PartToolResult, ToolResult: &result}
}

// Text joins the message's text parts.
func (m Message) Text() string {
	var text string
	for _, part := range m.Parts {
		if part.Kind == PartText {
			text += part.Text
		}
	}
	return text
}

// ToolCalls returns the tool calls the message makes.
func (m Message) ToolCalls() []ToolCall {
	var calls []ToolCall
	for _, part := range m.Parts {
		if part.Kind == PartToolCall && part.ToolCall != nil {
			calls = append(calls, *part.ToolCall)
		}
	}
	return calls
}
