package chat

import (
	"github.com/xmhrnx/giode"

	"gocode/ui/theme"
)

// queuedPanel shows messages waiting for the agent's next step.
func queuedPanel(c *ChatTab) giode.Element {
	texts := c.queue.texts()
	if len(texts) == 0 {
		return giode.Box(giode.Styles{Display: giode.DisplayNone})
	}
	rows := []giode.Element{giode.Text("Queued for the next step", giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Tiny,
		FontWeight: giode.FontWeightSemiBold, TextTransform: giode.TextTransformUppercase})}
	for _, text := range texts {
		rows = append(rows, giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			theme.Icon("clock", 13, theme.Muted),
			giode.Text(text, giode.Styles{Color: theme.Text, FontSize: theme.Sizes.Small, MaxLines: 2,
				TextOverflow: giode.TextOverflowEllipsis})))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 5, Padding: giode.SymmetricInset(14, 10), BorderRadius: 12,
		Background: theme.Panel, BorderWidth: 1, BorderColor: theme.SoftBorder}, rows...)
}
