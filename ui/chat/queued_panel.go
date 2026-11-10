package chat

import (
	"github.com/xmhrnx/giode"

	"gocode/ui/theme"
)

// queuedPanel shows messages waiting for the agent's next step, with a
// button that stops the current step and sends them at once.
type queuedPanel struct {
	tab     *ChatTab
	sendNow *giode.ButtonElement
}

func newQueuedPanel(tab *ChatTab) *queuedPanel {
	p := &queuedPanel{tab: tab}
	p.sendNow = giode.Button("Send now", giode.Styles{Background: theme.Primary, Color: theme.OnPrimary, BorderRadius: 7,
		FontSize: theme.Sizes.Small, Padding: giode.SymmetricInset(10, 4)}).OnClick(tab.sendQueuedNow)
	return p
}

func (p *queuedPanel) Layout(gtx giode.Context) giode.Dimensions {
	texts := p.tab.queue.texts()
	if len(texts) == 0 {
		return giode.Dimensions{}
	}
	header := giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
		giode.Text("Queued for the next step", giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Tiny, FlexGrow: 1,
			FontWeight: giode.FontWeightSemiBold, TextTransform: giode.TextTransformUppercase}),
		p.sendNow)
	rows := []giode.Element{header}
	for _, text := range texts {
		rows = append(rows, giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			theme.Icon("clock", 13, theme.Muted),
			giode.Text(text, giode.Styles{Color: theme.Text, FontSize: theme.Sizes.Small, MaxLines: 2,
				TextOverflow: giode.TextOverflowEllipsis})))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 5, Padding: giode.SymmetricInset(14, 10), BorderRadius: 12,
		Background: theme.Panel, BorderWidth: 1, BorderColor: theme.SoftBorder}, rows...).Layout(gtx)
}
