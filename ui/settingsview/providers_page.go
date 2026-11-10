package settingsview

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"
	"github.com/xmhrnx/giode/components/segmented"

	"gocode/internal/catalog"
	"gocode/internal/providers"
	"gocode/internal/secrets"
	"gocode/ui/core"
	"gocode/ui/theme"
)

// providersPage lists the providers of one kind at a time (labs, hosts,
// local servers, custom ones with the form that adds them), or every
// provider matching a search.
type providersPage struct {
	services *core.Services
	rows     map[string]*providerRow
	hasKey   map[string]bool
	testing  map[string]bool
	custom   *customProviderForm
	segment  *segmented.Control
	search   *input.Input
}

// providerSegment is one tab of the page: a category of providers.
type providerSegment struct {
	label, title, category string
}

// segments are the page's tabs, in order.
var segments = []providerSegment{
	{"Labs", "Model labs", providers.CategoryCloud},
	{"Hosts", "Routers and hosts", providers.CategoryRouter},
	{"Local", "On this computer", providers.CategoryLocal},
	{"Custom", "Custom", "custom"},
}

func newProvidersPage(s *core.Services) *providersPage {
	p := &providersPage{services: s, rows: map[string]*providerRow{},
		hasKey: map[string]bool{}, testing: map[string]bool{}}
	p.custom = newCustomProviderForm(p)
	labels := make([]string, len(segments))
	for i, seg := range segments {
		labels[i] = seg.label
	}
	p.segment = giode.Input.Segmented(labels...)
	p.search = giode.Input.Text("Search providers…", fieldStyle(-1))
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

func (p *providersPage) Layout(gtx giode.Context) giode.Dimensions {
	query := strings.ToLower(strings.TrimSpace(p.search.Text()))
	children := []giode.Element{
		giode.Row(giode.Styles{Width: -1, Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			p.segment, giode.Box(giode.Styles{FlexGrow: 1}, p.search)),
	}
	defs := catalog.Providers(p.services.Settings)
	shown := 0
	for i, seg := range segments {
		// A search looks through every segment; otherwise only the
		// chosen one shows.
		if query == "" && i != p.segment.Selected() {
			continue
		}
		rows := p.segmentRows(defs, seg.category, query)
		shown += len(rows)
		if seg.category == "custom" && query == "" {
			children = append(children, p.custom)
		}
		if len(rows) > 0 {
			children = append(children, groupTitle(seg.title), giode.Stack(giode.Styles{Width: -1, Gap: 2}, rows...))
		}
	}
	if shown == 0 && query != "" {
		children = append(children, giode.Text("No provider matches \u201c"+p.search.Text()+"\u201d.", theme.Caption()))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 10}, children...).Layout(gtx)
}

// segmentRows returns the rows of category's providers that match
// query: the ones in use first, then the rest by name.
func (p *providersPage) segmentRows(defs []providers.Definition, category, query string) []giode.Element {
	var matched []providers.Definition
	for _, def := range defs {
		if def.Category == category && (query == "" || matchesProvider(def, query)) {
			matched = append(matched, def)
		}
	}
	sort.SliceStable(matched, func(i, j int) bool {
		a, b := p.inUse(matched[i].ID), p.inUse(matched[j].ID)
		if a != b {
			return a
		}
		return strings.ToLower(matched[i].Name) < strings.ToLower(matched[j].Name)
	})
	rows := make([]giode.Element, len(matched))
	for i, def := range matched {
		rows[i] = p.row(def)
	}
	return rows
}

// inUse reports whether a provider is switched on or has a key.
func (p *providersPage) inUse(id string) bool {
	ps, ok := p.services.Settings.Provider(id)
	return p.hasKey[id] || (ok && ps.Enabled)
}

func matchesProvider(def providers.Definition, query string) bool {
	return strings.Contains(strings.ToLower(def.Name), query) || strings.Contains(strings.ToLower(def.ID), query) ||
		strings.Contains(strings.ToLower(def.BaseURL), query)
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
