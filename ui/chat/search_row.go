package chat

import (
	"github.com/xmhrnx/giode"

	"gocode/ui/theme"
)

// searchRow shows a search the provider ran itself, such as Claude's web
// search; its results are kept for the model, not shown.
type searchRow struct {
	rowBase
}

func newSearchRow(tab *ChatTab, it item) *searchRow { return &searchRow{rowBase{tab: tab, it: it}} }

func (r *searchRow) version() int { return 0 }

func (r *searchRow) Layout(gtx giode.Context) giode.Dimensions {
	part, _ := r.tab.part(r.it)
	return giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
		Padding: giode.SymmetricInset(0, 4)},
		theme.Icon("globe", 14, theme.Muted),
		giode.Text(part.Text, giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Small, FlexGrow: 1}),
	).Layout(gtx)
}
