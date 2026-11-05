package chat

import (
	"fmt"

	"github.com/xmhrnx/giode"

	"gocode/ui/theme"
)

// usageLabel shows the chat's total token usage, when enabled.
func usageLabel(c *ChatTab) giode.Element {
	usage := c.session.Usage
	if !c.services.Settings.Chat.ShowTokenUsage || usage.InputTokens+usage.OutputTokens == 0 {
		return giode.Box(giode.Styles{})
	}
	input := usage.InputTokens + usage.CacheReadTokens + usage.CacheWriteTokens
	text := fmt.Sprintf("%s in · %s out", compactNumber(input), compactNumber(usage.OutputTokens))
	if usage.CacheReadTokens > 0 {
		text += fmt.Sprintf(" · %d%% cached", usage.CacheReadTokens*100/max(input, 1))
	}
	return giode.Box(giode.Styles{Padding: giode.Inset{Left: 8}},
		giode.Text(text, giode.Styles{Color: theme.Faint, FontSize: theme.Sizes.Tiny}))
}

func compactNumber(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1000:
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}
	return fmt.Sprint(n)
}

// contextLabel shows how full the model's context window is.
func contextLabel(c *ChatTab) giode.Element {
	window := c.contextWindow()
	if c.contextTokens == 0 || window == 0 {
		return giode.Box(giode.Styles{})
	}
	percent := c.contextTokens * 100 / window
	color := theme.Faint
	if percent >= c.services.Settings.Agent.CompactAtPercent-10 {
		color = theme.Warning
	}
	return giode.Box(giode.Styles{Padding: giode.Inset{Left: 8}},
		giode.Text(fmt.Sprintf("context %d%%", percent), giode.Styles{Color: color, FontSize: theme.Sizes.Tiny}))
}
