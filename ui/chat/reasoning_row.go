package chat

import (
	"github.com/xmhrnx/giode"

	"gocode/ui/theme"
)

// reasoningRow shows one segment of the model's reasoning. The first
// segment carries the header, which opens or folds the whole part. It
// starts open, or folded with the "Keep reasoning collapsed" setting.
type reasoningRow struct {
	rowBase
	toggle *giode.PressableElement
}

func newReasoningRow(tab *ChatTab, it item) *reasoningRow {
	r := &reasoningRow{rowBase: rowBase{tab: tab, it: it}}
	r.toggle = giode.Pressable(r.header).Label("Toggle reasoning").
		OnClick(func(giode.ClickEvent) {
			group := r.it.group
			tab.toggledReasoning[group] = !tab.toggledReasoning[group]
		})
	return r
}

func (r *reasoningRow) collapsed() bool {
	return r.tab.services.Settings.Chat.CollapseReasoning != r.tab.toggledReasoning[r.it.group]
}

func (r *reasoningRow) version() int {
	if r.collapsed() {
		return 1
	}
	return 0
}

func (r *reasoningRow) Layout(gtx giode.Context) giode.Dimensions {
	if !r.tab.services.Settings.Chat.ShowReasoning || (r.it.segment > 0 && r.collapsed()) {
		return giode.Dimensions{}
	}
	// Indented under the header's label. Plain text rather than a panel:
	// the segments are separate rows and must read as one block.
	text := giode.Box(giode.Styles{Width: -1, Padding: giode.Inset{Left: 20}},
		giode.Text(r.it.text, giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Small,
			LineHeight: theme.Body().LineHeight - 3}))
	if r.it.segment > 0 {
		return text.Layout(gtx)
	}
	children := []giode.Element{r.toggle}
	if !r.collapsed() {
		children = append(children, text)
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 6, Padding: giode.Inset{Top: 4}}, children...).Layout(gtx)
}

func (r *reasoningRow) header(state giode.PressState) giode.Element {
	color := theme.Muted
	if state.Hovered {
		color = theme.Text
	}
	label, chevron := "Thought", "chevron-down"
	if r.it.draft {
		label = "Thinking…"
	}
	if r.collapsed() {
		chevron = "chevron-right"
	}
	return giode.Row(giode.Styles{Gap: 6, Align: giode.AlignCenter, Padding: giode.SymmetricInset(0, 2)},
		theme.Icon("brain", 14, color),
		giode.Text(label, giode.Styles{Color: color, FontSize: theme.Sizes.Small, FontWeight: giode.FontWeightMedium}),
		theme.Icon(chevron, 12, color),
	)
}
