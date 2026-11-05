package openai

import (
	"strings"

	"gocode/internal/llm"
)

// assistantMessage converts a stored reply. It reports false for a reply
// with neither text nor tool calls (e.g. cancelled before any output),
// which the API would reject.
func assistantMessage(msg llm.Message, echoField string) (chatMessage, bool) {
	var text, reasoning strings.Builder
	var calls []chatToolCall
	for _, part := range msg.Parts {
		switch {
		case part.Kind == llm.PartText:
			text.WriteString(part.Text)
		case part.Kind == llm.PartReasoning && !part.Redacted:
			reasoning.WriteString(part.Text)
		case part.Kind == llm.PartToolCall && part.ToolCall != nil:
			calls = append(calls, outgoingToolCall(*part.ToolCall))
		}
	}
	if text.Len() == 0 && len(calls) == 0 {
		return chatMessage{}, false
	}
	out := chatMessage{Role: "assistant", ToolCalls: calls}
	if text.Len() > 0 {
		out.Content = text.String()
	}
	if echoField != "" && reasoning.Len() > 0 {
		out.reasoningField, out.reasoning = echoField, reasoning.String()
	}
	return out, true
}

func outgoingToolCall(call llm.ToolCall) chatToolCall {
	arguments := string(call.Arguments)
	if strings.TrimSpace(arguments) == "" {
		arguments = "{}"
	}
	return chatToolCall{
		ID:       call.ID,
		Type:     "function",
		Function: chatFunctionCall{Name: call.Name, Arguments: arguments},
	}
}
