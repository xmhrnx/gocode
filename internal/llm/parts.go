package llm

// TextPart returns a text part.
func TextPart(text string) Part { return Part{Kind: PartText, Text: text} }

// ImagePart returns an image part.
func ImagePart(mediaType string, data []byte) Part {
	return Part{Kind: PartImage, Image: &Image{MediaType: mediaType, Data: data}}
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
