package gemini

import (
	"bytes"
	"encoding/json"
	"strings"

	"gocode/internal/llm"
)

// convertContents turns history into Gemini turns, dropping empty turns
// and merging consecutive turns by the same role.
func convertContents(history []llm.Message) []wireContent {
	var out []wireContent
	for _, msg := range history {
		role, parts := "user", userParts(msg)
		if msg.Role == llm.RoleAssistant {
			role, parts = "model", modelParts(msg)
		}
		if len(parts) == 0 {
			continue
		}
		if n := len(out); n > 0 && out[n-1].Role == role {
			out[n-1].Parts = append(out[n-1].Parts, parts...)
			continue
		}
		out = append(out, wireContent{Role: role, Parts: parts})
	}
	return out
}

// modelParts reproduces the reply's parts in order, each with the thought
// signature it arrived with, as the API requires.
func modelParts(msg llm.Message) []wirePart {
	var parts []wirePart
	for _, part := range msg.Parts {
		switch {
		case part.Kind == llm.PartReasoning && (part.Text != "" || part.Signature != ""):
			parts = append(parts, wirePart{Text: textPtr(part.Text), Thought: true, ThoughtSignature: part.Signature})
		case part.Kind == llm.PartText && (part.Text != "" || part.Signature != ""):
			parts = append(parts, wirePart{Text: textPtr(part.Text), ThoughtSignature: part.Signature})
		case part.Kind == llm.PartToolCall && part.ToolCall != nil:
			call := part.ToolCall
			parts = append(parts, wirePart{
				FunctionCall:     &functionCall{ID: providerID(call.ID), Name: call.Name, Args: objectOrEmpty(call.Arguments)},
				ThoughtSignature: call.Signature,
			})
		case part.Kind == llm.PartProvider && part.Provider == providerName && len(part.Raw) > 0:
			parts = append(parts, wirePart{raw: part.Raw})
		case part.Kind == llm.PartMedia && part.Media != nil && part.Media.IsImage():
			// Images go back so the model can edit them in a later turn.
			parts = append(parts, wirePart{InlineData: &inlineData{MimeType: part.Media.MediaType, Data: part.Media.Data},
				ThoughtSignature: part.Signature})
		}
	}
	return parts
}

// inlineTypes are the documents Gemini takes inline, with the type it
// names each by; others are sent as their text or description.
var inlineTypes = map[string]string{
	"application/pdf": "application/pdf",
	"audio/wav":       "audio/wav", "audio/mpeg": "audio/mp3", "audio/aiff": "audio/aiff",
	"audio/aac": "audio/aac", "audio/ogg": "audio/ogg", "audio/flac": "audio/flac",
	"video/mp4": "video/mp4", "video/mpeg": "video/mpeg", "video/quicktime": "video/mov",
	"video/avi": "video/avi", "video/x-flv": "video/x-flv", "video/webm": "video/webm",
	"video/x-ms-wmv": "video/wmv", "video/3gpp": "video/3gpp",
}

// userParts puts function responses first, answering the previous turn's
// calls, then the user's text and images.
func userParts(msg llm.Message) []wirePart {
	var responses, rest []wirePart
	for _, part := range msg.Parts {
		switch {
		case part.Kind == llm.PartToolResult && part.ToolResult != nil:
			responses = append(responses, wirePart{FunctionResponse: toFunctionResponse(*part.ToolResult)})
		case part.Kind == llm.PartText && part.Text != "":
			rest = append(rest, wirePart{Text: textPtr(part.Text)})
		case part.Kind == llm.PartImage && part.Image != nil:
			rest = append(rest, wirePart{InlineData: &inlineData{MimeType: part.Image.MediaType, Data: part.Image.Data}})
		case part.Kind == llm.PartDocument && part.Document != nil:
			if mime, ok := inlineTypes[part.Document.MediaType]; ok {
				rest = append(rest, wirePart{InlineData: &inlineData{MimeType: mime, Data: part.Document.Data}})
			} else {
				rest = append(rest, wirePart{Text: textPtr(llm.DocumentText(*part.Document))})
			}
		}
	}
	return append(responses, rest...)
}

func toFunctionResponse(result llm.ToolResult) *functionResponse {
	key := "content"
	if result.IsError {
		key = "error"
	}
	return &functionResponse{
		ID:       providerID(result.CallID),
		Name:     result.Name,
		Response: map[string]any{key: result.Content},
	}
}

// providerID drops ids this adapter invented for calls that came without
// one, so the API never sees ids it didn't issue.
func providerID(id string) string {
	if strings.HasPrefix(id, generatedIDPrefix) {
		return ""
	}
	return id
}

// objectOrEmpty returns arguments when they are a JSON object, else {}:
// args must be an object.
func objectOrEmpty(arguments json.RawMessage) json.RawMessage {
	trimmed := bytes.TrimSpace(arguments)
	if len(trimmed) > 0 && trimmed[0] == '{' && json.Valid(trimmed) {
		return trimmed
	}
	return json.RawMessage("{}")
}
