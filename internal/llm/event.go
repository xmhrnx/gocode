package llm

// EventKind identifies a streaming event.
type EventKind int

// Event is one streamed piece of a reply.
type Event struct {
	Kind      EventKind
	Text      string
	Signature string
	// Redacted marks a reasoning event whose Signature is an encrypted
	// reasoning payload rather than a signature over streamed text.
	Redacted bool
	ToolCall *ToolCall
	// Part is a complete provider part, for EventPart.
	Part       *Part
	Usage      *Usage
	StopReason string
}

// Usage counts tokens for one call.
type Usage struct {
	InputTokens      int `json:"input_tokens"`
	OutputTokens     int `json:"output_tokens"`
	CacheReadTokens  int `json:"cache_read_tokens,omitempty"`
	CacheWriteTokens int `json:"cache_write_tokens,omitempty"`
}

const (
	// EventText carries a chunk of reply text in Text.
	EventText EventKind = iota
	// EventReasoning carries a chunk of reasoning in Text, or, at the end
	// of a reasoning block, its Signature.
	EventReasoning
	// EventToolCall carries one complete tool call.
	EventToolCall
	// EventPart carries a complete PartProvider or PartMedia part in Part.
	EventPart
	// EventUsage reports token usage; it may arrive more than once, each
	// time with the totals so far.
	EventUsage
	// EventStop reports why the model stopped, in StopReason.
	EventStop
)

// Add returns the sum of two usages.
func (u Usage) Add(other Usage) Usage {
	return Usage{
		InputTokens:      u.InputTokens + other.InputTokens,
		OutputTokens:     u.OutputTokens + other.OutputTokens,
		CacheReadTokens:  u.CacheReadTokens + other.CacheReadTokens,
		CacheWriteTokens: u.CacheWriteTokens + other.CacheWriteTokens,
	}
}
