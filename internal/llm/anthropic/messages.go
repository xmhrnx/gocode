package anthropic

import (
	"bytes"
	"encoding/base64"
	"encoding/json"

	"gocode/internal/llm"
)

// convertMessages turns history into Messages API turns. Empty turns are
// dropped and consecutive turns by the same role merged, since the API
// rejects both.
func convertMessages(history []llm.Message) []wireMessage {
	var out []wireMessage
	for _, msg := range history {
		role, blocks := string(msg.Role), userBlocks(msg)
		if msg.Role == llm.RoleAssistant {
			blocks = assistantBlocks(msg)
		}
		if len(blocks) == 0 {
			continue
		}
		if n := len(out); n > 0 && out[n-1].Role == role {
			out[n-1].Content = toolResultsFirst(append(out[n-1].Content, blocks...))
			continue
		}
		out = append(out, wireMessage{Role: role, Content: blocks})
	}
	return out
}

// assistantBlocks reproduces the reply's blocks in their original order;
// thinking blocks must come back exactly as received.
func assistantBlocks(msg llm.Message) []any {
	var blocks []any
	for _, part := range msg.Parts {
		switch {
		case part.Kind == llm.PartReasoning && part.Redacted:
			blocks = append(blocks, redactedThinkingBlock{Type: "redacted_thinking", Data: part.Signature})
		case part.Kind == llm.PartReasoning && part.Signature != "":
			blocks = append(blocks, thinkingBlock{Type: "thinking", Thinking: part.Text, Signature: part.Signature})
		case part.Kind == llm.PartText && part.Text != "":
			blocks = append(blocks, textBlock{Type: "text", Text: part.Text})
		case part.Kind == llm.PartToolCall && part.ToolCall != nil:
			call := part.ToolCall
			blocks = append(blocks, toolUseBlock{Type: "tool_use", ID: call.ID, Name: call.Name, Input: objectOrEmpty(call.Arguments)})
		case part.Kind == llm.PartProvider && part.Provider == providerName && len(part.Raw) > 0:
			blocks = append(blocks, part.Raw)
		}
	}
	return blocks
}

// userBlocks puts tool results first, as the API requires them to lead
// the turn that follows a tool_use.
func userBlocks(msg llm.Message) []any {
	var blocks []any
	for _, part := range msg.Parts {
		switch {
		case part.Kind == llm.PartToolResult && part.ToolResult != nil:
			result := part.ToolResult
			blocks = append(blocks, toolResultBlock{Type: "tool_result", ToolUseID: result.CallID, Content: result.Content, IsError: result.IsError})
		case part.Kind == llm.PartText && part.Text != "":
			blocks = append(blocks, textBlock{Type: "text", Text: part.Text})
		case part.Kind == llm.PartDocument && part.Document != nil && part.Document.MediaType != "application/pdf":
			// Claude takes PDFs, not audio or video: those are described.
			blocks = append(blocks, textBlock{Type: "text", Text: llm.DocumentText(*part.Document)})
		case part.Kind == llm.PartDocument && part.Document != nil:
			doc := part.Document
			blocks = append(blocks, documentBlock{Type: "document", Title: doc.Name, Source: imageSource{
				Type: "base64", MediaType: doc.MediaType, Data: base64.StdEncoding.EncodeToString(doc.Data),
			}})
		case part.Kind == llm.PartImage && part.Image != nil:
			blocks = append(blocks, imageBlock{Type: "image", Source: imageSource{
				Type: "base64", MediaType: part.Image.MediaType, Data: base64.StdEncoding.EncodeToString(part.Image.Data),
			}})
		}
	}
	return toolResultsFirst(blocks)
}

// toolResultsFirst stably moves tool_result blocks to the front.
func toolResultsFirst(blocks []any) []any {
	ordered := make([]any, 0, len(blocks))
	for _, block := range blocks {
		if _, ok := block.(toolResultBlock); ok {
			ordered = append(ordered, block)
		}
	}
	for _, block := range blocks {
		if _, ok := block.(toolResultBlock); !ok {
			ordered = append(ordered, block)
		}
	}
	return ordered
}

// objectOrEmpty returns arguments when they are a JSON object, else {}:
// tool_use input must be an object.
func objectOrEmpty(arguments json.RawMessage) json.RawMessage {
	trimmed := bytes.TrimSpace(arguments)
	if len(trimmed) > 0 && trimmed[0] == '{' && json.Valid(trimmed) {
		return trimmed
	}
	return json.RawMessage("{}")
}
