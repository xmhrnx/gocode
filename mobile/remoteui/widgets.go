package remoteui

import (
	"github.com/xmhrnx/giode"

	"gocode/ui/components"
	"gocode/ui/theme"
)

// screen lays a screen out: a header with an optional back button, then
// the body filling the rest.
func screen(back *components.IconButton, title, subtitle string, trailing giode.Element, body ...giode.Element) giode.Element {
	head := []giode.Element{}
	if back != nil {
		head = append(head, back)
	}
	texts := []giode.Element{giode.Text(title, giode.Styles{Color: theme.Text, FontSize: theme.Sizes.Heading,
		FontWeight: giode.FontWeightSemiBold, MaxLines: 1})}
	if subtitle != "" {
		texts = append(texts, giode.Text(subtitle, giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Small, MaxLines: 1}))
	}
	head = append(head, giode.Stack(giode.Styles{Gap: 2, FlexGrow: 1}, texts...))
	if trailing != nil {
		head = append(head, trailing)
	}
	children := []giode.Element{
		giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
			Padding: giode.Inset{Left: 12, Right: 12, Top: 14, Bottom: 10}}, head...),
		giode.Divider(giode.Styles{Color: theme.SoftBorder}),
	}
	children = append(children, body...)
	return giode.Stack(giode.Styles{Width: -1, Height: -1}, children...)
}

// primaryButton is a filled button.
func primaryButton(label string, onClick func()) *giode.ButtonElement {
	return giode.Button(label, giode.Styles{Background: theme.Primary, Color: theme.OnPrimary, BorderRadius: 10,
		Padding: giode.SymmetricInset(18, 12)}).OnClick(onClick)
}

// wideButton is a filled button as wide as its column.
func wideButton(label string, onClick func()) *giode.ButtonElement {
	return giode.Button(label, giode.Styles{Width: -1, Background: theme.Primary, Color: theme.OnPrimary, BorderRadius: 10,
		Padding: giode.SymmetricInset(18, 12)}).OnClick(onClick)
}

// outlineButton is a bordered button.
func outlineButton(label string, onClick func()) *giode.ButtonElement {
	return giode.Button(label, giode.Styles{BorderColor: theme.Border, Color: theme.Text, BorderRadius: 10,
		Padding: giode.SymmetricInset(18, 12)}).Variant(giode.VariantOutlined).OnClick(onClick)
}

// note is a centered, muted message filling the body.
func note(icon, title, text string) giode.Element {
	return giode.Stack(giode.Styles{Width: -1, FlexGrow: 1, Justify: giode.JustifyCenter, Align: giode.AlignCenter,
		Gap: 8, Padding: giode.SymmetricInset(28, 20)},
		theme.Icon(icon, 28, theme.Faint),
		giode.Text(title, giode.Styles{Color: theme.Text, FontWeight: giode.FontWeightMedium, TextAlign: giode.TextAlignCenter}),
		giode.Text(text, giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Small, TextAlign: giode.TextAlignCenter}),
	)
}

// backButton goes to the previous screen.
func backButton(a *App) *components.IconButton {
	return components.NewIconButton("chevron-left", "Back", func() { a.router.Back() }).SetSize(20)
}
