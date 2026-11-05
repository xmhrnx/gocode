package agent

import (
	"errors"
	"strings"

	"gocode/internal/llm"
)

// overflowPhrases appear in providers' errors for a request that doesn't
// fit the context window. Wording varies; these cover OpenAI, Anthropic,
// Gemini, vLLM, llama.cpp, LM Studio and the routers in front of them.
var overflowPhrases = []string{
	"context length", "context window", "maximum context", "context_length",
	"prompt is too long", "too many tokens", "input is too long",
	"exceeds the limit", "reduce the length", "token limit", "too large for model",
}

// isContextOverflow reports whether err says the request was too long.
func isContextOverflow(err error) bool {
	var apiErr *llm.APIError
	if !errors.As(err, &apiErr) || (apiErr.StatusCode != 400 && apiErr.StatusCode != 413) {
		return false
	}
	message := strings.ToLower(apiErr.Message)
	for _, phrase := range overflowPhrases {
		if strings.Contains(message, phrase) {
			return true
		}
	}
	return false
}
