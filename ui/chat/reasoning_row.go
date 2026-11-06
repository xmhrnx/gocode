package chat

import (
	"github.com/xmhrnx/giode"

	"gocode/ui/theme"
)

// reasoningRow shows the model's reasoning, collapsed to one line unless
// the user opens it or it is still streaming.
type reasoningRow struct {
	rowBase
	expanded bool
	toggle   *giode.PressableElement
}

func newReasoningRow(tab *ChatTab, it item) *reasoningRow {
	r := &reasoningRow{rowBase: rowBase{tab: tab, it: it}}
	r.toggle = giode.Pressable(r.header).Label("Toggle reasoning").
		OnClick(func(giode.ClickEvent) { r.expanded = !r.expanded })
	return r
}

func (r *reasoningRow) version() int {
	if r.expanded {
		return 1
	}
	return 0
}

func (r *reasoningRow) Layout(gtx giode.Context) giode.Dimensions {
	if !r.tab.services.Settings.Chat.ShowReasoning {
		return giode.Dimensions{}
	}
	part, _ := r.tab.part(r.it)
	children := []giode.Element{r.toggle}
	if r.expanded || r.it.draft {
		text := part.Text
		if r.it.draft {
			text = liveTail(text, maxLiveReasoning)
		}
		// A panel rather than a rule stretched to the text's height: a
		// stretched sibling makes the text be measured and then laid out
		// again, doubling the cost of every streamed frame.
		children = append(children, giode.Box(giode.Styles{
			Width: -1, Padding: giode.SymmetricInset(12, 8), BorderRadius: 8, Background: theme.Raised,
		}, giode.Text(text, giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Small,
			LineHeight: theme.Body().LineHeight - 3})))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 6, Padding: giode.SymmetricInset(0, 4)}, children...).Layout(gtx)
}

func (r *reasoningRow) header(state giode.PressState) giode.Element {
	color := theme.Muted
	if state.Hovered {
		color = theme.Text
	}
	label, chevron := "Thought", "chevron-right"
	if r.it.draft {
		label = "Thinking…"
	}
	if r.expanded || r.it.draft {
		chevron = "chevron-down"
	}
	return giode.Row(giode.Styles{Gap: 6, Align: giode.AlignCenter, Padding: giode.SymmetricInset(0, 2)},
		theme.Icon("brain", 14, color),
		giode.Text(label, giode.Styles{Color: color, FontSize: theme.Sizes.Small, FontWeight: giode.FontWeightMedium}),
		theme.Icon(chevron, 12, color),
	)
}
