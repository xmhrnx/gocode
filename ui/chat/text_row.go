package chat

import (
	"context"

	"gioui.org/layout"
	"github.com/xmhrnx/giode"

	"gocode/ui/components"
	"gocode/ui/theme"
)

// textRow shows assistant text as Markdown. At the end of a turn it adds
// copy and retry actions.
type textRow struct {
	rowBase
	copy  giode.Element
	retry *components.IconButton
}

func newTextRow(tab *ChatTab, it item) *textRow {
	r := &textRow{rowBase: rowBase{tab: tab, it: it}}
	r.copy = giode.CopyButton(func() string { return tab.turnText(r.it.message) }, giode.Styles{Color: theme.Muted})
	r.retry = components.NewIconButton("rotate-cw", "Retry", tab.retry).SetSize(14)
	return r
}

func (r *textRow) version() int {
	if r.it.turnEnd {
		return 1
	}
	return 0
}

func (r *textRow) Layout(gtx layout.Context) layout.Dimensions {
	part, _ := r.tab.part(r.it)
	if isBlank(part.Text) {
		return layout.Dimensions{}
	}
	children := r.content(part.Text)
	if r.it.turnEnd && r.tab.run == nil {
		children = append(children, giode.Row(giode.Styles{Gap: 2, Align: giode.AlignCenter, Margin: giode.Inset{Left: -6}},
			r.copy, r.retry))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 6, Padding: giode.SymmetricInset(0, 6)}, children...).Layout(gtx)
}

// content renders the text. While it streams, the settled blocks and the
// growing tail are separate documents, so only the tail is parsed again.
func (r *textRow) content(text string) []giode.Element {
	if !r.it.draft {
		return []giode.Element{markdown(r.tab, text, r.it.key)}
	}
	settled, tail := splitSettled(text)
	if settled == "" {
		return []giode.Element{markdown(r.tab, tail, r.it.key+"/tail")}
	}
	return []giode.Element{markdown(r.tab, settled, r.it.key+"/settled"), markdown(r.tab, tail, r.it.key+"/tail")}
}

// markdown renders assistant text in the chat's style.
func markdown(tab *ChatTab, text, key string) giode.Element {
	return giode.Markdown(text, giode.MarkdownOptions{
		Styles:         theme.Body(),
		CodeFont:       "JetBrains Mono",
		CodeBackground: theme.CodeBack,
		Selectable:     key,
		OnLink: func(url string) {
			go func() {
				if err := tab.services.App.OpenURL(context.Background(), url); err != nil {
					tab.services.Error("Couldn't open the link: " + err.Error())
				}
			}()
		},
	})
}
