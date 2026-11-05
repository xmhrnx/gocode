package settingsview

import (
	"context"
	"fmt"

	"gioui.org/layout"
	"github.com/xmhrnx/giode"

	"gocode/internal/catalog"
	"gocode/internal/providers"
	"gocode/internal/secrets"
	"gocode/ui/core"
	"gocode/ui/theme"
)

// providersPage lists every provider grouped by kind, with the form that
// adds a custom one.
type providersPage struct {
	services *core.Services
	rows     map[string]*providerRow
	hasKey   map[string]bool
	testing  map[string]bool
	custom   *customProviderForm
}

func newProvidersPage(s *core.Services) *providersPage {
	p := &providersPage{services: s, rows: map[string]*providerRow{},
		hasKey: map[string]bool{}, testing: map[string]bool{}}
	p.custom = newCustomProviderForm(p)
	p.loadKeyStatus()
	return p
}

// loadKeyStatus checks which providers have a stored key, off the UI
// goroutine since the keyring may block.
func (p *providersPage) loadKeyStatus() {
	defs := catalog.Providers(p.services.Settings)
	keys, app := p.services.Keys, p.services.App
	go func() {
		found := map[string]bool{}
		for _, def := range defs {
			_, err := keys.Get(secrets.ProviderKeyName(def.ID))
			found[def.ID] = err == nil
		}
		app.Post(func() { p.hasKey = found })
	}()
}

func (p *providersPage) find(id string) (providers.Definition, bool) {
	return catalog.Find(p.services.Settings, id)
}

func (p *providersPage) Layout(gtx layout.Context) layout.Dimensions {
	groups := []struct{ title, category string }{
		{"Model labs", providers.CategoryCloud}, {"Routers and hosts", providers.CategoryRouter},
		{"On this computer", providers.CategoryLocal}, {"Custom", "custom"},
	}
	children := []giode.Element{p.custom}
	defs := catalog.Providers(p.services.Settings)
	for _, g := range groups {
		var rows []giode.Element
		for _, def := range defs {
			if def.Category == g.category {
				rows = append(rows, p.row(def))
			}
		}
		if len(rows) > 0 {
			children = append(children, groupTitle(g.title), giode.Stack(giode.Styles{Width: -1, Gap: 2}, rows...))
		}
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 10}, children...).Layout(gtx)
}

// row returns the provider's row, creating it on first use.
func (p *providersPage) row(def providers.Definition) *providerRow {
	row, ok := p.rows[def.ID]
	if !ok {
		row = newProviderRow(p, def)
		p.rows[def.ID] = row
	}
	return row
}

func groupTitle(title string) giode.Element {
	return giode.Box(giode.Styles{Padding: giode.Inset{Top: 10, Left: 4}},
		giode.Text(title, giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Tiny,
			FontWeight: giode.FontWeightSemiBold, TextTransform: giode.TextTransformUppercase}))
}

// keyBadge says whether the provider is ready to use.
func (p *providersPage) keyBadge(def providers.Definition) giode.Element {
	label, color := "No key", theme.Muted
	switch {
	case p.hasKey[def.ID]:
		label, color = "Key saved", theme.Success
	case def.KeyOptional:
		label = "No key needed"
	}
	return giode.Text(label, giode.Styles{Color: color, FontSize: theme.Sizes.Tiny})
}

// testStatus reports the result of the last connection test.
func (p *providersPage) testStatus(id string) giode.Element {
	status := p.services.Models.Status(id)
	text, color := "", theme.Muted
	switch {
	case status.Loading:
		text = "Connecting…"
	case status.Err != nil:
		text, color = status.Err.Error(), theme.Danger
	case p.testing[id]:
		text, color = fmt.Sprintf("Connected · %d models", len(status.Models)), theme.Success
	}
	return giode.Text(text, giode.Styles{Color: color, FontSize: theme.Sizes.Small, MaxLines: 2})
}

func (p *providersPage) test(id string) {
	if def, ok := p.find(id); ok {
		p.testing[id] = true
		p.services.Models.Refresh(def)
	}
}

func (p *providersPage) openURL(url string) {
	go func() {
		if err := p.services.App.OpenURL(context.Background(), url); err != nil {
			p.services.Error("Couldn't open the link: " + err.Error())
		}
	}()
}
