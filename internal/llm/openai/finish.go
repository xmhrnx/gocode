package openai

import "gocode/internal/llm"

// chunkUsage is the token count sent in the last chunk when
// stream_options.include_usage is set.
type chunkUsage struct {
	PromptTokens        int `json:"prompt_tokens"`
	CompletionTokens    int `json:"completion_tokens"`
	PromptTokensDetails *struct {
		CachedTokens int `json:"cached_tokens"`
	} `json:"prompt_tokens_details"`
	// PromptCacheHitTokens is DeepSeek's name for cached input.
	PromptCacheHitTokens int `json:"prompt_cache_hit_tokens"`
}

// finish flushes buffered output once the stream has ended.
func (s *streamState) finish() error {
	if !s.sawDone && s.finishReason == "" {
		return errIncomplete
	}
	s.think.flush(s.emitSplit)
	calls := s.calls.complete()
	for i := range calls {
		s.emit(llm.Event{Kind: llm.EventToolCall, ToolCall: &calls[i]})
	}
	s.emit(llm.Event{Kind: llm.EventStop, StopReason: stopReason(s.finishReason, len(calls) > 0)})
	return nil
}

// toLLM reports uncached input separately from cache reads, matching how
// Anthropic counts them.
func (u chunkUsage) toLLM() *llm.Usage {
	cached := u.PromptCacheHitTokens
	if u.PromptTokensDetails != nil {
		cached = max(cached, u.PromptTokensDetails.CachedTokens)
	}
	return &llm.Usage{
		InputTokens:     u.PromptTokens - cached,
		OutputTokens:    u.CompletionTokens,
		CacheReadTokens: cached,
	}
}

// stopReason maps finish reasons to the Anthropic-style names used across
// adapters. Some servers say "stop" even when calling tools.
func stopReason(finish string, calledTools bool) string {
	switch {
	case calledTools || finish == "tool_calls" || finish == "function_call":
		return "tool_use"
	case finish == "stop" || finish == "":
		return "end_turn"
	case finish == "length":
		return "max_tokens"
	}
	return finish
}
