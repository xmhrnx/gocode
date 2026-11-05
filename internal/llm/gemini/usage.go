package gemini

import "gocode/internal/llm"

// usageMetadata is the running token count sent with response chunks.
type usageMetadata struct {
	PromptTokenCount        int `json:"promptTokenCount"`
	CandidatesTokenCount    int `json:"candidatesTokenCount"`
	ThoughtsTokenCount      int `json:"thoughtsTokenCount"`
	CachedContentTokenCount int `json:"cachedContentTokenCount"`
}

// toLLM reports uncached input separately from cache reads, matching how
// Anthropic counts them, and bills thinking as output.
func (u usageMetadata) toLLM() *llm.Usage {
	return &llm.Usage{
		InputTokens:     u.PromptTokenCount - u.CachedContentTokenCount,
		OutputTokens:    u.CandidatesTokenCount + u.ThoughtsTokenCount,
		CacheReadTokens: u.CachedContentTokenCount,
	}
}
