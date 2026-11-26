package remoteui

import (
	"github.com/xmhrnx/giode"

	"gocode/internal/remote/wire"
	"gocode/ui/theme"
)

// reasoningShown bounds the reasoning a phone shows of each row.
const reasoningShown = 600

// rowView draws one transcript row.
func rowView(r wire.Row) giode.Element {
	switch r.Kind {
	case wire.RowUser:
		return giode.Stack(giode.Styles{Width: -1, Align: giode.AlignEnd, Padding: giode.Inset{Top: 10}},
			giode.Box(giode.Styles{Padding: giode.SymmetricInset(12, 9), BorderRadius: 14, Background: theme.Raised,
				BorderWidth: 1, BorderColor: theme.SoftBorder, MaxWidth: 520},
				giode.Text(r.Text, giode.Styles{Color: theme.Text})))
	case wire.RowText:
		return giode.Markdown(r.Text, giode.MarkdownOptions{Styles: theme.Body(), CodeFont: "JetBrains Mono",
			CodeBackground: theme.CodeBack, Math: true})
	case wire.RowReasoning:
		text := r.Text
		if runes := []rune(text); len(runes) > reasoningShown {
			text = "…" + string(runes[len(runes)-reasoningShown:])
		}
		return giode.Row(giode.Styles{Gap: 8, Wrap: giode.WrapNoWrap},
			theme.Icon("brain", 14, theme.Faint),
			giode.Text(text, giode.Styles{Color: theme.Faint, FontSize: theme.Sizes.Small, FlexGrow: 1}))
	case wire.RowTool:
		return toolRow(r)
	case wire.RowMedia:
		return quietRow("image", r.Text+" (open it on the desktop)")
	}
	return quietRow("info", r.Text)
}

// toolRow shows a tool call with its state.
func toolRow(r wire.Row) giode.Element {
	state := giode.Element(giode.Spinner(giode.Styles{Width: 12, Height: 12, Color: theme.Primary}))
	switch r.State {
	case wire.StateDone:
		state = theme.Icon("circle-check", 14, theme.Success)
	case wire.StateFailed:
		state = theme.Icon("circle-x", 14, theme.Danger)
	}
	return giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
		Padding: giode.SymmetricInset(10, 7), BorderRadius: 8, Background: theme.Panel},
		state,
		giode.Text(r.Text, giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Small, FlexGrow: 1, MaxLines: 2}))
}

func quietRow(icon, text string) giode.Element {
	return giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
		theme.Icon(icon, 14, theme.Muted),
		giode.Text(text, giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Small, FlexGrow: 1}))
}
