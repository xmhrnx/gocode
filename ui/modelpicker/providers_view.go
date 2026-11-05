package modelpicker

import (
	"fmt"

	"github.com/xmhrnx/giode"

	"gocode/internal/catalog"
	"gocode/internal/providers"
	"gocode/ui/theme"
)

// providersView lists recent models and the enabled providers.
func (p *Picker) providersView() giode.Element {
	query := p.search.Text()
	p.firstMatch = nil
	var items []giode.Element
	if recent := p.recentRows(query); len(recent) > 0 {
		items = append(append(items, sectionTitle("Recent")), recent...)
	}
	var rows []giode.Element
	for _, def := range catalog.Providers(p.services.Settings) {
		if !catalog.Enabled(p.services.Settings, def.ID) || !matches(query, def.Name) {
			continue
		}
		if p.firstMatch == nil {
			id := def.ID
			p.firstMatch = func() { p.showProvider(id) }
		}
		rows = append(rows, p.providerRow(def))
	}
	if len(rows) > 0 {
		items = append(append(items, sectionTitle("Providers")), rows...)
	}
	if len(items) == 0 {
		items = append(items, message(noProvidersText(query)))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 1, Padding: giode.Inset{Right: 8}}, items...)
}

// recentRows lists recently used models that are still available.
func (p *Picker) recentRows(query string) []giode.Element {
	var rows []giode.Element
	models := &p.services.Settings.Models
	for _, ref := range models.Recent {
		parsed := catalog.ParseRef(ref)
		def, ok := catalog.Find(p.services.Settings, parsed.Provider)
		if !ok || !catalog.Enabled(p.services.Settings, def.ID) || models.IsHidden(ref) || !matches(query, def.Name+" "+parsed.Model) {
			continue
		}
		rows = append(rows, p.modelRow("recent:"+ref, ref, parsed.Model, def.Name))
	}
	return rows
}

func (p *Picker) providerRow(def providers.Definition) giode.Element {
	return p.providerRows.Get(def.ID, func() *giode.PressableElement {
		return giode.Pressable(func(s giode.PressState) giode.Element {
			return providerRowContent(def.Name, p.providerDetail(def.ID), s)
		}).Label(def.Name).OnClick(func(giode.ClickEvent) { p.showProvider(def.ID) })
	})
}

// providerDetail describes the provider's model list: its size, or why
// it isn't there.
func (p *Picker) providerDetail(id string) string {
	status := p.services.Models.Status(id)
	def, _ := catalog.Find(p.services.Settings, id)
	switch {
	case status.Loading:
		return "loading…"
	case status.Err != nil:
		return "couldn't list models"
	}
	return fmt.Sprintf("%d models", len(p.services.Models.Models(def)))
}

func providerRowContent(name, detail string, s giode.PressState) giode.Element {
	background := ""
	if s.Hovered || s.Focused {
		background = theme.Hover
	}
	return giode.Row(giode.Styles{Width: -1, Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
		Padding: giode.SymmetricInset(10, 9), BorderRadius: 8, Background: background},
		theme.Icon("plug", 15, theme.Muted),
		giode.Text(name, giode.Styles{Color: theme.Text, FlexGrow: 1}),
		giode.Text(detail, giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Small}),
		theme.Icon("chevron-right", 14, theme.Muted),
	)
}

func noProvidersText(query string) string {
	if query != "" {
		return "No providers match “" + query + "”."
	}
	return "No providers are enabled yet. Add a key or a local server in Settings → Providers."
}
