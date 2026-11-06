package chat

import (
	"fmt"

	"github.com/xmhrnx/giode"

	"gocode/ui/theme"
)

// taskProgress returns the progress of a running task's sub-agent, or nil.
func (c *ChatTab) taskProgress(callID string) *subagentProgress {
	if c.run == nil {
		return nil
	}
	return c.run.subagents[callID]
}

// subagentView lists a sub-agent's latest steps inside its task's card.
func subagentView(p *subagentProgress) giode.Element {
	rows := make([]giode.Element, 0, len(p.activity)+1)
	for _, step := range p.activity {
		rows = append(rows, giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			theme.Icon("corner-down-left", 12, theme.Faint),
			giode.Text(step, giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Small, FontFamily: "JetBrains Mono",
				TextOverflow: giode.TextOverflowEllipsis})))
	}
	status := fmt.Sprintf("%d steps", p.steps)
	if p.steps == 1 {
		status = "1 step"
	}
	if p.status != "" {
		status += " · " + p.status
	}
	rows = append(rows, giode.Text(status, giode.Styles{Color: theme.Faint, FontSize: theme.Sizes.Tiny}))
	return giode.Stack(giode.Styles{Width: -1, Gap: 4, Padding: giode.SymmetricInset(14, 8)}, rows...)
}
