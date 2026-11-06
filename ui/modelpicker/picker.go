// Package modelpicker is the model chooser shown from every chat: recent
// models and the connected providers first, then a provider's models,
// searchable, one level down.
package modelpicker

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/dialog"
	"github.com/xmhrnx/giode/components/input"
	"github.com/xmhrnx/giode/components/scroll"

	"gocode/internal/catalog"
	"gocode/ui/core"
	"gocode/ui/theme"
)

// Picker is a dialog choosing a model; choosing one calls onSelect with
// its "provider/model" reference.
type Picker struct {
	// Current, when set, returns the selected model's reference, which is
	// marked in the lists.
	Current func() string

	services *core.Services
	onSelect func(ref string)
	open     bool
	// providerID is the provider whose models are shown; "" shows the
	// list of providers.
	providerID string
	// firstMatch is what Enter in the search field picks.
	firstMatch func()

	dialog       *dialog.Dialog
	search       *input.Input
	list         *scroll.Scroll
	back         *giode.PressableElement
	manage       *giode.ButtonElement
	modelRows    *giode.KeyedCache[string, *giode.PressableElement]
	providerRows *giode.KeyedCache[string, *giode.PressableElement]
}

// New returns a closed picker.
func New(services *core.Services, onSelect func(ref string)) *Picker {
	p := &Picker{services: services, onSelect: onSelect,
		modelRows:    giode.NewKeyedCache[string, *giode.PressableElement](),
		providerRows: giode.NewKeyedCache[string, *giode.PressableElement]()}
	p.search = giode.Input.Text("Search providers…", giode.Styles{
		Background: theme.Canvas, BorderColor: theme.Border, BorderRadius: 9, Color: theme.Text,
	}).Submit(func(string) {
		if p.firstMatch != nil {
			p.firstMatch()
		}
	})
	p.list = giode.Scroll(p.listView, giode.Styles{ScrollBar: giode.ScrollBar{Width: 6, AutoHide: true, TrackColor: "transparent"}})
	p.back = giode.Pressable(backButton).Label("Back to providers").OnClick(func(giode.ClickEvent) { p.showProvider("") })
	p.manage = giode.Button("Manage providers", giode.Styles{Color: theme.Text, Background: theme.Hover,
		FontSize: theme.Sizes.Small, BorderRadius: 8, Padding: giode.SymmetricInset(12, 6)}).
		Variant(giode.VariantGhost).OnClick(func() {
		p.open = false
		services.Open().OpenSettings("Providers")
	})
	p.dialog = giode.Dialog(&p.open, p.content, giode.Styles{
		Background: theme.Panel, BorderRadius: 14, BorderWidth: 1, BorderColor: theme.Border,
		BoxShadow: giode.BoxShadow{Y: 18, Blur: 48, Color: "#00000066"},
	}).Padding(giode.UniformInset(14))
	return p
}

// Open shows the provider list and refreshes model lists not fetched yet.
func (p *Picker) Open() {
	p.open = true
	p.showProvider("")
	settings := p.services.Settings
	for _, def := range catalog.Providers(settings) {
		status := p.services.Models.Status(def.ID)
		if catalog.Enabled(settings, def.ID) && status.Models == nil && !status.Loading {
			p.services.Models.Refresh(def)
		}
	}
}

// Layout lays out the dialog; it takes no space.
func (p *Picker) Layout(gtx giode.Context) giode.Dimensions { return p.dialog.Layout(gtx) }

// showProvider switches to a provider's models, or to the provider list
// for "".
func (p *Picker) showProvider(id string) {
	p.providerID = id
	p.search.SetText("")
	p.search.Placeholder("Search providers…")
	if def, ok := catalog.Find(p.services.Settings, id); ok {
		p.search.Placeholder("Search " + def.Name + " models…")
	}
	p.search.Focus()
	p.list.ScrollTo(0)
}

func (p *Picker) choose(ref string) {
	p.open = false
	p.onSelect(ref)
}

func (p *Picker) current() string {
	if p.Current == nil {
		return ""
	}
	return p.Current()
}
