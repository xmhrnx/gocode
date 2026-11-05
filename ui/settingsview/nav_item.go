package settingsview

import (
	"github.com/xmhrnx/giode"

	"gocode/ui/theme"
)

// newNavItem returns the navigation button for a category.
func newNavItem(t *Tab, c category) giode.Element {
	return giode.Pressable(func(state giode.PressState) giode.Element {
		active := t.active == c.name && t.search.Text() == ""
		background, color := "", theme.Muted
		switch {
		case active:
			background, color = theme.PrimarySoft, theme.Text
		case state.Hovered:
			background, color = theme.Hover, theme.Text
		}
		iconColor := color
		if active {
			iconColor = theme.Primary
		}
		return giode.Row(giode.Styles{Width: -1, Gap: 10, Align: giode.AlignCenter, Padding: giode.SymmetricInset(10, 7),
			BorderRadius: 8, Background: background},
			theme.Icon(c.icon, 15, iconColor),
			giode.Text(c.name, giode.Styles{Color: color, FontSize: theme.Sizes.Small, FontWeight: giode.FontWeightMedium}))
	}).Label(c.name).OnClick(func(giode.ClickEvent) { t.ShowCategory(c.name) })
}
