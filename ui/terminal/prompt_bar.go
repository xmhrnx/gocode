package terminal

import (
	"github.com/xmhrnx/giode"

	"gocode/ui/theme"
)

// promptBar is the command input at the bottom of the terminal.
func (t *Tab) promptBar() giode.Element {
	return giode.Box(giode.Styles{Width: -1, Padding: giode.UniformInset(14)},
		giode.Stack(giode.Styles{Width: -1, Gap: 6, Padding: giode.SymmetricInset(14, 10), BorderRadius: 12,
			Background: theme.Panel, BorderWidth: 1, BorderColor: theme.Border},
			giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter},
				theme.Icon("folder", 13, theme.Muted),
				giode.Text(shortDir(t.session.WorkDir()), giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Small,
					FontFamily: "JetBrains Mono"}),
			),
			completionList(t.matches),
			giode.Row(giode.Styles{Width: -1, Gap: 10, Align: giode.AlignStart, Wrap: giode.WrapNoWrap},
				giode.Text("❯", giode.Styles{Color: theme.Primary, FontFamily: "JetBrains Mono", FontWeight: giode.FontWeightBold}),
				giode.Box(giode.Styles{FlexGrow: 1}, t.input),
			),
		),
	)
}

// welcome is shown before the first command.
func welcome() giode.Element {
	return giode.Stack(giode.Styles{Width: -1, Height: -1, Align: giode.AlignCenter, Justify: giode.JustifyCenter, Gap: 10},
		theme.Icon("square-terminal", 32, theme.Faint),
		giode.Text("Each command runs as its own block.", giode.Styles{Color: theme.Muted}),
		giode.Text("Tab completes · ↑/↓ recall · Esc stops · Mod+L clears · ✦ asks the agent about a block",
			giode.Styles{Color: theme.Faint, FontSize: theme.Sizes.Small}),
	)
}
