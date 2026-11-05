package settingsview

import (
	"github.com/xmhrnx/giode"

	"gocode/ui/theme"
)

// page builds the right side: the active category, or search results
// from every category.
func (t *Tab) page() giode.Element {
	query := t.search.Text()
	var sections []giode.Element
	for _, c := range categories {
		var rows []giode.Element
		for _, e := range t.entries {
			if e.category == c.name && t.shows(e, query) {
				rows = append(rows, entryView(e))
			}
		}
		if len(rows) > 0 {
			sections = append(sections, section(c.name, rows))
		}
	}
	if len(sections) == 0 {
		sections = append(sections, giode.Text("No settings match “"+query+"”.", giode.Styles{Color: theme.Muted}))
	}
	return giode.Row(giode.Styles{Width: -1, Justify: giode.JustifyCenter, Padding: giode.SymmetricInset(28, 24)},
		giode.Stack(giode.Styles{Width: -1, MaxWidth: 860, Gap: 28}, sections...))
}

func (t *Tab) shows(e entry, query string) bool {
	if query != "" {
		return e.matches(query)
	}
	return e.category == t.active
}

func section(title string, rows []giode.Element) giode.Element {
	separated := make([]giode.Element, 0, len(rows)*2)
	for i, row := range rows {
		if i > 0 {
			separated = append(separated, giode.Divider(giode.Styles{Color: theme.SoftBorder}))
		}
		separated = append(separated, row)
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 12},
		giode.Text(title, giode.Styles{Color: theme.Text, FontSize: theme.Sizes.Heading, FontWeight: giode.FontWeightSemiBold}),
		giode.Stack(giode.Styles{Width: -1, BorderRadius: 12, Background: theme.Panel, BorderWidth: 1,
			BorderColor: theme.SoftBorder, Padding: giode.SymmetricInset(18, 4)}, separated...),
	)
}

func entryView(e entry) giode.Element {
	text := giode.Stack(giode.Styles{Gap: 3, FlexGrow: 1},
		giode.Text(e.title, giode.Styles{Color: theme.Text, FontWeight: giode.FontWeightMedium}),
		giode.Text(e.description, giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Small}),
	)
	if e.description == "" {
		text = giode.Stack(giode.Styles{FlexGrow: 1}, giode.Text(e.title, giode.Styles{Color: theme.Text}))
	}
	if e.title == e.category {
		text = giode.Stack(giode.Styles{FlexGrow: 1}, giode.Text(e.description, giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Small}))
	}
	if e.wide {
		return giode.Stack(giode.Styles{Width: -1, Gap: 12, Padding: giode.SymmetricInset(0, 14)}, text, e.control)
	}
	return giode.Row(giode.Styles{Width: -1, Gap: 24, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
		Padding: giode.SymmetricInset(0, 14)}, text, e.control)
}
