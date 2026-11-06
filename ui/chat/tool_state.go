package chat

import "gocode/internal/llm"

// toolVersion summarizes a tool call's state, so its row rebuilds when
// the call starts running or finishes.
func (c *ChatTab) toolVersion(callID string) int {
	switch {
	case c.results[callID] != nil:
		return 0
	case c.toolResult(callID) != nil:
		return 1
	case c.run != nil && c.run.running[callID]:
		if progress, ok := c.run.subagents[callID]; ok {
			return 4 + progress.revision // a task's card shows its sub-agent's progress
		}
		return 2
	}
	return 3
}

// toolResult returns the result of a tool call, or nil if it has none
// yet, from the index built with the transcript items.
func (c *ChatTab) toolResult(callID string) *llm.ToolResult {
	if result, ok := c.results[callID]; ok {
		return result
	}
	if c.run != nil {
		if result, ok := c.run.finished[callID]; ok {
			return &result
		}
	}
	return nil
}

// indexToolResults maps call IDs to their results.
func indexToolResults(messages []llm.Message) map[string]*llm.ToolResult {
	results := map[string]*llm.ToolResult{}
	for _, msg := range messages {
		for _, part := range msg.Parts {
			if part.Kind == llm.PartToolResult {
				results[part.ToolResult.CallID] = part.ToolResult
			}
		}
	}
	return results
}

// toolRunning reports whether the call is executing now.
func (c *ChatTab) toolRunning(callID string) bool {
	return c.run != nil && c.run.running[callID]
}

// part returns the message part an item shows.
func (c *ChatTab) part(it item) (llm.Part, bool) {
	msg := c.message(it)
	if it.part < 0 || it.part >= len(msg.Parts) {
		return llm.Part{}, false
	}
	return msg.Parts[it.part], true
}

// message returns the message an item belongs to.
func (c *ChatTab) message(it item) llm.Message {
	if it.draft {
		if c.run == nil {
			return llm.Message{}
		}
		return c.run.draftMessage()
	}
	if it.message >= len(c.session.Messages) {
		return llm.Message{}
	}
	return c.session.Messages[it.message]
}
