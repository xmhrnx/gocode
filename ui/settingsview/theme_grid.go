package settingsview

import (
	"github.com/xmhrnx/giode"

	"gocode/ui/core"
	"gocode/ui/theme"
)

// themeGrid shows a card per theme; clicking one applies it.
type themeGrid struct {
	cards []giode.Element
}

func newThemeGrid(s *core.Services) *themeGrid {
	g := &themeGrid{}
	g.cards = append(g.cards, newThemeCard(s, theme.SystemID, "System", theme.Resolve(theme.SystemID)))
	for _, t := range theme.All() {
		g.cards = append(g.cards, newThemeCard(s, t.ID, t.Name, t))
	}
	return g
}

func (g *themeGrid) Layout(gtx giode.Context) giode.Dimensions {
	return giode.Row(giode.Styles{Width: -1, Gap: 10}, g.cards...).Layout(gtx)
}

// newThemeCard returns a card previewing a theme with its own colors.
func newThemeCard(s *core.Services, id, name string, preview theme.Theme) giode.Element {
	return giode.Pressable(func(state giode.PressState) giode.Element {
		selected := s.Settings.Appearance.Theme == id
		border := theme.SoftBorder
		if selected {
			border = theme.Primary
		} else if state.Hovered {
			border = theme.Border
		}
		p := preview.Palette
		return giode.Stack(giode.Styles{Width: 132, Gap: 8, Padding: giode.UniformInset(6), BorderRadius: 12,
			BorderWidth: 2, BorderColor: border},
			giode.Stack(giode.Styles{Width: -1, Height: 64, BorderRadius: 8, Background: p.Background,
				Padding: giode.UniformInset(8), Gap: 6},
				giode.Box(giode.Styles{Width: 70, Height: 10, BorderRadius: 4, Background: p.Surface}),
				giode.Box(giode.Styles{Width: 46, Height: 10, BorderRadius: 4, Background: p.Muted}),
				giode.Row(giode.Styles{Gap: 4},
					giode.Box(giode.Styles{Width: 24, Height: 10, BorderRadius: 4, Background: p.Primary}),
					giode.Box(giode.Styles{Width: 14, Height: 10, BorderRadius: 4, Background: p.Accent})),
			),
			giode.Text(name, giode.Styles{Color: theme.Text, FontSize: theme.Sizes.Small, FontWeight: giode.FontWeightMedium}),
		)
	}).Label(name + " theme").OnClick(func(giode.ClickEvent) {
		s.Settings.Appearance.Theme = id
		s.SettingsChanged()
	})
}
