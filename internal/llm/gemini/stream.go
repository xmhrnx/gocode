package gemini

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"gocode/internal/llm"
	"gocode/internal/llm/httpapi"
	"gocode/internal/llm/sse"
)

// streamResponse is one streamed GenerateContentResponse, or an error.
type streamResponse struct {
	Candidates []struct {
		Content      wireContent `json:"content"`
		FinishReason string      `json:"finishReason"`
	} `json:"candidates"`
	UsageMetadata  *usageMetadata `json:"usageMetadata"`
	PromptFeedback *struct {
		BlockReason string `json:"blockReason"`
	} `json:"promptFeedback"`
	Error *struct {
		Code int `json:"code"`
	} `json:"error"`
}

// streamState turns response chunks into llm events. Function calls arrive
// whole, so they are emitted as soon as they are seen.
type streamState struct {
	emit         func(llm.Event)
	callCount    int
	finishReason string
}

// generatedIDPrefix marks call ids invented for function calls that came
// without one.
const generatedIDPrefix = "gocode_call_"

// errIncomplete reports a stream that closed before a finish reason.
var errIncomplete = errors.New("stream ended before the reply was complete")

func newStreamState(emit func(llm.Event)) *streamState {
	return &streamState{emit: emit}
}

func (s *streamState) handle(event sse.Event) error {
	data := []byte(strings.TrimSpace(event.Data))
	if len(data) == 0 {
		return nil
	}
	var chunk streamResponse
	if err := json.Unmarshal(data, &chunk); err != nil {
		return fmt.Errorf("decode stream chunk: %w", err)
	}
	if chunk.Error != nil {
		return &llm.APIError{StatusCode: errorStatus(chunk.Error.Code), Message: httpapi.ErrorMessage(data, "stream error")}
	}
	if chunk.PromptFeedback != nil && chunk.PromptFeedback.BlockReason != "" {
		return &llm.APIError{StatusCode: http.StatusBadRequest, Message: "prompt blocked: " + chunk.PromptFeedback.BlockReason}
	}
	if len(chunk.Candidates) > 0 {
		candidate := chunk.Candidates[0]
		for _, part := range candidate.Content.Parts {
			s.handlePart(part)
		}
		if candidate.FinishReason != "" {
			s.finishReason = candidate.FinishReason
		}
	}
	if chunk.UsageMetadata != nil {
		s.emit(llm.Event{Kind: llm.EventUsage, Usage: chunk.UsageMetadata.toLLM()})
	}
	return nil
}

func (s *streamState) handlePart(part wirePart) {
	switch {
	case part.FunctionCall != nil:
		s.emit(llm.Event{Kind: llm.EventToolCall, ToolCall: s.toolCall(part)})
	case part.Thought:
		s.emitText(llm.EventReasoning, part)
	case part.Text != nil:
		s.emitText(llm.EventText, part)
	}
}

// emitText sends text with its thought signature; a part may carry only a
// signature, which must still be kept.
func (s *streamState) emitText(kind llm.EventKind, part wirePart) {
	if part.text() != "" || part.ThoughtSignature != "" {
		s.emit(llm.Event{Kind: kind, Text: part.text(), Signature: part.ThoughtSignature})
	}
}

func (s *streamState) toolCall(part wirePart) *llm.ToolCall {
	s.callCount++
	call := part.FunctionCall
	id := call.ID
	if id == "" {
		id = fmt.Sprintf("%s%d", generatedIDPrefix, s.callCount)
	}
	return &llm.ToolCall{ID: id, Name: call.Name, Arguments: objectOrEmpty(call.Args), Signature: part.ThoughtSignature}
}

func (s *streamState) finish() error {
	if s.finishReason == "" {
		return errIncomplete
	}
	s.emit(llm.Event{Kind: llm.EventStop, StopReason: stopReason(s.finishReason, s.callCount > 0)})
	return nil
}

// stopReason maps finish reasons to the Anthropic-style names used across
// adapters; Gemini reports STOP even after function calls.
func stopReason(finish string, calledTools bool) string {
	switch {
	case calledTools:
		return "tool_use"
	case finish == "STOP":
		return "end_turn"
	case finish == "MAX_TOKENS":
		return "max_tokens"
	}
	return strings.ToLower(finish)
}

// errorStatus treats a missing code in a mid-stream error as a server
// failure.
func errorStatus(code int) int {
	if code < 400 || code > 599 {
		return http.StatusInternalServerError
	}
	return code
}
