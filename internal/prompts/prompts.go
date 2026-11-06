// Package prompts builds the system prompts sent to the model.
package prompts

import (
	_ "embed"
	"strings"
)

var (
	//go:embed defaults/agent.md
	defaultAgent string
	//go:embed defaults/chat.md
	defaultChat string
	//go:embed defaults/title.md
	defaultTitle string
	//go:embed defaults/compact.md
	defaultCompact string
	//go:embed defaults/subagent.md
	defaultSubagent string
)

// DefaultAgent returns the built-in agent prompt.
func DefaultAgent() string { return defaultAgent }

// DefaultChat returns the built-in prompt for chats without tools.
func DefaultChat() string { return defaultChat }

// DefaultTitle returns the built-in prompt that names conversations.
func DefaultTitle() string { return defaultTitle }

// DefaultCompact returns the built-in prompt that summarizes a
// conversation when its context is compacted.
func DefaultCompact() string { return defaultCompact }

// DefaultSubagent returns the section added to a sub-agent's prompt.
func DefaultSubagent() string { return defaultSubagent }

// Or returns custom unless it is blank, in which case it returns fallback.
func Or(custom, fallback string) string {
	if strings.TrimSpace(custom) == "" {
		return fallback
	}
	return custom
}
