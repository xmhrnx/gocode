package agent

import (
	"context"
	"fmt"
	"strings"

	"gocode/internal/llm"
)

// charsPerToken estimates text size in tokens without a tokenizer; it is
// only used to size the summarizing request.
const charsPerToken = 4

// needsCompaction reports whether the last request used enough of the
// context window that the next one may not fit.
func (c Config) needsCompaction(lastInputTokens int) bool {
	if c.CompactAtPercent <= 0 || c.ContextWindow <= 0 {
		return false
	}
	return lastInputTokens*100 >= c.ContextWindow*c.CompactAtPercent
}

// compact asks the model to summarize history and returns the history to
// continue with: one message holding the summary.
func compact(ctx context.Context, cfg Config, history []llm.Message) ([]llm.Message, string, error) {
	budget := max(cfg.ContextWindow*charsPerToken/2, 20000)
	req := llm.Request{
		Model: cfg.Model, System: cfg.CompactPrompt, ReasoningEffort: "",
		Messages: []llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{
			llm.TextPart("Conversation to summarize:\n" + transcriptText(history, budget)),
		}}},
	}
	var summary strings.Builder
	err := cfg.Client.Stream(ctx, req, func(ev llm.Event) {
		if ev.Kind == llm.EventText {
			summary.WriteString(ev.Text)
		}
	})
	if err != nil {
		return history, "", fmt.Errorf("compact the conversation: %w", err)
	}
	text := strings.TrimSpace(summary.String())
	if text == "" {
		return history, "", fmt.Errorf("compact the conversation: the model returned an empty summary")
	}
	return []llm.Message{SummaryMessage(text)}, text, nil
}

// SummaryMessage is the message that stands in for a compacted
// conversation.
func SummaryMessage(summary string) llm.Message {
	return llm.Message{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart(
		"The conversation so far was compacted to save context. Summary of everything before this point:\n\n" +
			summary + "\n\nContinue the task from where it left off, without asking the user to repeat anything.",
	)}}
}

// Compact summarizes history on request, outside a run.
func Compact(ctx context.Context, cfg Config, history []llm.Message) (string, error) {
	_, summary, err := compact(ctx, cfg, history)
	return summary, err
}
