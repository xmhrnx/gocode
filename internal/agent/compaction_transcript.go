package agent

import (
	"fmt"
	"strings"

	"gocode/internal/llm"
)

// maxToolResultInSummary bounds each tool result in the text sent for
// summarizing: the summary needs outcomes, not whole file dumps.
const maxToolResultInSummary = 1500

// transcriptText renders the conversation as plain text for the
// summarizing call. Plain text sidesteps every provider rule about tool
// call pairing, and lets the text be cut to fit.
func transcriptText(history []llm.Message, budgetChars int) string {
	var b strings.Builder
	for _, msg := range history {
		for _, part := range msg.Parts {
			writePart(&b, msg.Role, part)
		}
	}
	return fitBudget(b.String(), budgetChars)
}

func writePart(b *strings.Builder, role llm.Role, part llm.Part) {
	switch part.Kind {
	case llm.PartText:
		fmt.Fprintf(b, "\n[%s]\n%s\n", role, part.Text)
	case llm.PartImage:
		fmt.Fprintf(b, "\n[%s attached an image]\n", role)
	case llm.PartToolCall:
		fmt.Fprintf(b, "\n[tool call %s %s]\n", part.ToolCall.Name, clip(string(part.ToolCall.Arguments), maxToolResultInSummary))
	case llm.PartToolResult:
		status := "result"
		if part.ToolResult.IsError {
			status = "error"
		}
		fmt.Fprintf(b, "[tool %s]\n%s\n", status, clip(part.ToolResult.Content, maxToolResultInSummary))
	}
}

// fitBudget keeps the start (the task) and the end (the latest work) of
// text when it is longer than budget.
func fitBudget(text string, budget int) string {
	if budget <= 0 || len(text) <= budget {
		return text
	}
	head := budget / 5
	tail := budget - head
	return text[:head] + "\n\n[… middle of the conversation omitted …]\n\n" + text[len(text)-tail:]
}

func clip(text string, limit int) string {
	if len(text) <= limit {
		return text
	}
	return text[:limit] + " …[cut]"
}
