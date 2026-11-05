package session

import (
	"testing"

	"gocode/internal/llm"
)

func TestModelHistoryUsesTheSummaryAfterCompaction(t *testing.T) {
	sess := New("/", "p/m")
	text := func(s string) llm.Message { return llm.Message{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart(s)}} }
	sess.Messages = append(sess.Messages, text("old"))
	sess.Compacted("sum")
	sess.Messages = append(sess.Messages, text("new"))
	history := sess.ModelHistory(func(s string) llm.Message { return text("S:" + s) })
	if len(history) != 2 || history[0].Text() != "S:sum" || history[1].Text() != "new" {
		t.Fatalf("history = %+v", history)
	}
}
