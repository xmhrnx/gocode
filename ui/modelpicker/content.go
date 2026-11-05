package modelpicker

import (
	"github.com/xmhrnx/giode"

	"gocode/ui/theme"
)

func (p *Picker) content() giode.Element {
	header := giode.Element(p.search)
	hint := "Enter opens the first provider · Esc closes"
	if p.providerID != "" {
		header = giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			p.back, giode.Box(giode.Styles{FlexGrow: 1}, p.search))
		hint = "Enter picks the first match · Esc closes"
	}
	return giode.Stack(giode.Styles{Width: 560, Gap: 10},
		header,
		giode.Box(giode.Styles{Width: -1, Height: 420}, p.list),
		giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Text(hint, giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Tiny}),
			giode.Box(giode.Styles{FlexGrow: 1}),
			p.manage,
		),
	)
}

// listView builds the provider list or the open provider's models.
func (p *Picker) listView() giode.Element {
	if p.providerID != "" {
		return p.modelsView()
	}
	return p.providersView()
}

func sectionTitle(title string) giode.Element {
	return giode.Box(giode.Styles{Padding: giode.Inset{Top: 12, Bottom: 4, Left: 8}},
		giode.Text(title, giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Tiny,
			FontWeight: giode.FontWeightSemiBold, TextTransform: giode.TextTransformUppercase}))
}

func message(text string) giode.Element {
	return giode.Box(giode.Styles{Width: -1, Padding: giode.UniformInset(24), Justify: giode.JustifyCenter},
		giode.Text(text, giode.Styles{Color: theme.Muted, TextAlign: giode.TextAlignCenter}))
}

func backButton(state giode.PressState) giode.Element {
	background := ""
	if state.Hovered {
		background = theme.Hover
	}
	return giode.Box(giode.Styles{Padding: giode.UniformInset(8), BorderRadius: 8, Background: background},
		theme.Icon("chevron-left", 16, theme.Text))
}
