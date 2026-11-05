package anthropic

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gocode/internal/llm"
	"gocode/internal/llm/sse"
)

// streamState turns Messages API events into llm events.
type streamState struct {
	emit       func(llm.Event)
	toolUses   map[int]*pendingToolUse
	usage      llm.Usage
	stopReason string
	stopped    bool
}

// pendingToolUse collects a tool_use block whose input arrives as JSON
// fragments.
type pendingToolUse struct {
	id, name string
	input    strings.Builder
}

// errIncomplete reports a stream that closed before message_stop.
var errIncomplete = errors.New("stream ended before the reply was complete")

func newStreamState(emit func(llm.Event)) *streamState {
	return &streamState{emit: emit, toolUses: map[int]*pendingToolUse{}}
}

func (s *streamState) handle(event sse.Event) error {
	if strings.TrimSpace(event.Data) == "" {
		return nil
	}
	var payload streamEvent
	if err := json.Unmarshal([]byte(event.Data), &payload); err != nil {
		return fmt.Errorf("decode %s event: %w", event.Name, err)
	}
	switch payload.Type {
	case "message_start":
		s.updateUsage(payload.Message.Usage)
	case "content_block_start":
		s.startBlock(payload.Index, payload.ContentBlock)
	case "content_block_delta":
		s.applyDelta(payload.Index, payload.Delta)
	case "content_block_stop":
		s.stopBlock(payload.Index)
	case "message_delta":
		s.stopReason = payload.Delta.StopReason
		s.updateUsage(payload.Usage)
	case "message_stop":
		s.stopped = true
		s.emit(llm.Event{Kind: llm.EventStop, StopReason: s.stopReason})
	case "error":
		return payload.Error.apiError()
	}
	return nil // ping and unknown future events need no handling
}

func (s *streamState) startBlock(index int, block wireBlock) {
	switch block.Type {
	case "text":
		s.emitText(llm.EventText, block.Text)
	case "thinking":
		s.emitText(llm.EventReasoning, block.Thinking)
	case "redacted_thinking":
		s.emit(llm.Event{Kind: llm.EventReasoning, Signature: block.Data, Redacted: true})
	case "tool_use":
		s.toolUses[index] = &pendingToolUse{id: block.ID, name: block.Name}
	}
}

func (s *streamState) applyDelta(index int, delta wireDelta) {
	switch delta.Type {
	case "text_delta":
		s.emitText(llm.EventText, delta.Text)
	case "thinking_delta":
		s.emitText(llm.EventReasoning, delta.Thinking)
	case "signature_delta":
		s.emit(llm.Event{Kind: llm.EventReasoning, Signature: delta.Signature})
	case "input_json_delta":
		if pending, ok := s.toolUses[index]; ok {
			pending.input.WriteString(delta.PartialJSON)
		}
	}
}

// stopBlock emits a finished tool call. Input cut short (e.g. by
// max_tokens) is not valid JSON and can't be run or sent back, so that
// call is dropped.
func (s *streamState) stopBlock(index int) {
	pending, ok := s.toolUses[index]
	if !ok {
		return
	}
	delete(s.toolUses, index)
	input := strings.TrimSpace(pending.input.String())
	if input == "" {
		input = "{}"
	}
	if !json.Valid([]byte(input)) {
		return
	}
	call := &llm.ToolCall{ID: pending.id, Name: pending.name, Arguments: json.RawMessage(input)}
	s.emit(llm.Event{Kind: llm.EventToolCall, ToolCall: call})
}

func (s *streamState) emitText(kind llm.EventKind, text string) {
	if text != "" {
		s.emit(llm.Event{Kind: kind, Text: text})
	}
}

// updateUsage merges counts (later events repeat or extend earlier ones)
// and reports the running totals.
func (s *streamState) updateUsage(usage *wireUsage) {
	if usage == nil {
		return
	}
	s.usage.InputTokens = max(s.usage.InputTokens, usage.InputTokens)
	s.usage.OutputTokens = max(s.usage.OutputTokens, usage.OutputTokens)
	s.usage.CacheReadTokens = max(s.usage.CacheReadTokens, usage.CacheReadInputTokens)
	s.usage.CacheWriteTokens = max(s.usage.CacheWriteTokens, usage.CacheCreationInputTokens)
	snapshot := s.usage
	s.emit(llm.Event{Kind: llm.EventUsage, Usage: &snapshot})
}

func (s *streamState) finish() error {
	if !s.stopped {
		return errIncomplete
	}
	return nil
}
