package agent

import (
	"gocode/internal/llm"
	"gocode/internal/tools"
)

// Config is what one run needs.
type Config struct {
	Client llm.Client
	Model  string
	System string
	// Tools is nil for a plain chat without tools.
	Tools       *tools.Registry
	ToolEnabled func(tools.Tool) bool
	Env         tools.Env
	Permissions *Permissions
	Ask         Asker

	MaxSteps        int
	MaxTokens       int
	Temperature     *float64
	ReasoningEffort string
	MaxRetries      int

	// ContextWindow is the model's context size in tokens; CompactAtPercent
	// of it triggers compaction (0 turns automatic compaction off).
	ContextWindow    int
	CompactAtPercent int
	CompactPrompt    string
}

func (c Config) request(history []llm.Message) llm.Request {
	req := llm.Request{
		Model: c.Model, System: c.System, Messages: history,
		MaxTokens: c.MaxTokens, Temperature: c.Temperature,
		ReasoningEffort: c.ReasoningEffort,
	}
	if c.Tools != nil {
		req.Tools = c.Tools.Specs(c.ToolEnabled)
	}
	return req
}
