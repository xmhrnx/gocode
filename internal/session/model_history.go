package session

import (
	"slices"

	"gocode/internal/llm"
)

// ModelHistory returns the conversation as the model should see it: after
// a compaction, summary stands in for the messages before ContextStart.
func (s *Session) ModelHistory(summary func(string) llm.Message) []llm.Message {
	if s.ContextSummary == "" || s.ContextStart > len(s.Messages) {
		return slices.Clone(s.Messages)
	}
	return append([]llm.Message{summary(s.ContextSummary)}, s.Messages[s.ContextStart:]...)
}

// Compacted records that everything so far is replaced by summary.
func (s *Session) Compacted(summary string) {
	s.ContextSummary, s.ContextStart = summary, len(s.Messages)
}
