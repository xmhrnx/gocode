package chat

import (
	"fmt"
	"math"
	"time"

	"github.com/xmhrnx/giode"

	"gocode/internal/llm"
	"gocode/ui/theme"
)

// turnStats sums the timing of the model calls in the turn that ends at
// message index: the turn's length, from its first call to its last, and
// its output speed over the time the model was writing.
func (c *ChatTab) turnStats(index int) (took time.Duration, perSecond float64, ok bool) {
	var first, last time.Time
	var tokens int
	var writing time.Duration
	for i := c.lastUserMessage(index+1) + 1; i <= index && i < len(c.session.Messages); i++ {
		stats := c.session.Messages[i].Stats
		if stats == nil || c.session.Messages[i].Role != llm.RoleAssistant {
			continue
		}
		if first.IsZero() {
			first = stats.Started
		}
		last = stats.Ended
		tokens += stats.OutputTokens
		if !stats.FirstOutput.IsZero() {
			writing += stats.Ended.Sub(stats.FirstOutput)
		}
	}
	if first.IsZero() {
		return 0, 0, false
	}
	if tokens > 0 && writing > 200*time.Millisecond {
		perSecond = float64(tokens) / writing.Seconds()
	}
	return last.Sub(first), perSecond, true
}

// turnStatsLabel shows a turn's length and speed, or nothing for turns
// saved before GoCode timed them.
func turnStatsLabel(c *ChatTab, index int) giode.Element {
	took, perSecond, ok := c.turnStats(index)
	if !ok {
		return giode.Box(giode.Styles{})
	}
	text := formatDuration(took)
	if perSecond > 0 {
		text += fmt.Sprintf(" · %.0f tok/s", math.Round(perSecond))
	}
	return giode.Box(giode.Styles{Padding: giode.Inset{Left: 6}},
		giode.Text(text, giode.Styles{Color: theme.Faint, FontSize: theme.Sizes.Tiny}))
}

// formatDuration shows a duration in seconds, or minutes and seconds.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1f s", d.Seconds())
	}
	return fmt.Sprintf("%d min %d s", int(d.Minutes()), int(d.Seconds())%60)
}
