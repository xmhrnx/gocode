package openai

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gocode/internal/llm"
	"gocode/internal/llm/sse"
)

// chatChunk is one streamed chat.completion.chunk, or an error payload.
type chatChunk struct {
	Choices []chunkChoice   `json:"choices"`
	Usage   *chunkUsage     `json:"usage"`
	Error   json.RawMessage `json:"error"`
}

type chunkChoice struct {
	Delta        chunkDelta `json:"delta"`
	FinishReason string     `json:"finish_reason"`
}

// streamState turns chunks into llm events. Tool calls are held until the
// stream ends because their arguments arrive in fragments.
type streamState struct {
	emit         func(llm.Event)
	think        thinkSplitter
	calls        toolCallAccumulator
	finishReason string
	sawDone      bool
	mediaCount   int
}

// errIncomplete reports a stream that closed before the reply finished.
var errIncomplete = errors.New("stream ended before the reply was complete")

func newStreamState(emit func(llm.Event)) *streamState {
	return &streamState{emit: emit}
}

func (s *streamState) handle(event sse.Event) error {
	data := strings.TrimSpace(event.Data)
	switch data {
	case "":
		return nil
	case "[DONE]":
		s.sawDone = true
		return sse.ErrStop
	}
	var chunk chatChunk
	if err := json.Unmarshal([]byte(data), &chunk); err != nil {
		return fmt.Errorf("decode stream chunk: %w", err)
	}
	if len(chunk.Error) > 0 && string(chunk.Error) != "null" {
		return streamError([]byte(data), chunk.Error)
	}
	if len(chunk.Choices) > 0 {
		choice := chunk.Choices[0]
		s.handleDelta(choice.Delta)
		if choice.FinishReason != "" {
			s.finishReason = choice.FinishReason
		}
	}
	if chunk.Usage != nil {
		s.emit(llm.Event{Kind: llm.EventUsage, Usage: chunk.Usage.toLLM()})
	}
	return nil
}

func (s *streamState) handleDelta(delta chunkDelta) {
	if reasoning := delta.reasoningText(); reasoning != "" {
		s.emit(llm.Event{Kind: llm.EventReasoning, Text: reasoning})
	}
	for _, piece := range contentPieces(delta.Content) {
		switch {
		case piece.imageURL != "":
			s.emitImage(piece.imageURL)
		case piece.text == "":
		case piece.reasoning:
			s.emit(llm.Event{Kind: llm.EventReasoning, Text: piece.text})
		default:
			s.think.feed(piece.text, s.emitSplit)
		}
	}
	for _, image := range delta.Images {
		s.emitImage(image.ImageURL.URL)
	}
	for _, call := range delta.ToolCalls {
		s.calls.add(call)
	}
}

// emitImage sends an image the model made. One given as a link instead
// of data is shown as that link.
func (s *streamState) emitImage(url string) {
	mediaType, data, ok := llm.ParseDataURL(url)
	if !ok {
		if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
			s.emit(llm.Event{Kind: llm.EventText, Text: "\n\n[Image](" + url + ")\n\n"})
		}
		return
	}
	s.mediaCount++
	media := llm.MediaPart(mediaType, data, s.mediaCount)
	s.emit(llm.Event{Kind: llm.EventPart, Part: &media})
}

func (s *streamState) emitSplit(reasoning bool, text string) {
	if text == "" {
		return
	}
	kind := llm.EventText
	if reasoning {
		kind = llm.EventReasoning
	}
	s.emit(llm.Event{Kind: kind, Text: text})
}
