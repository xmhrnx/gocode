package settingsview

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/checkbox"
	"github.com/xmhrnx/giode/components/input"

	"gocode/internal/catalog"
	"gocode/ui/core"
	"gocode/ui/theme"

	"gocode/ui/components"
)

// modelsPage chooses which models the picker offers: a list of the
// enabled providers, and for the open one its models, searchable, each
// with a checkbox.
type modelsPage struct {
	services *core.Services
	// providerID is the provider whose models are shown; "" lists providers.
	providerID string
	search     *input.Input
	back       *giode.ButtonElement
	refresh    *giode.ButtonElement
	providers  map[string]*giode.PressableElement
	boxes      *giode.KeyedCache[string, *checkbox.Checkbox]
}

func newModelsPage(s *core.Services) *modelsPage {
	p := &modelsPage{services: s, providers: map[string]*giode.PressableElement{},
		boxes: giode.NewKeyedCache[string, *checkbox.Checkbox]()}
	p.search = giode.Input.Text("Search models…", fieldStyle(-1))
	components.AddEditMenu(s.App, p.search, components.EditOptions{})
	p.back = smallButton("All providers", false).Icon(theme.Icon("chevron-left", 14, "")).OnClick(func() { p.open("") })
	p.refresh = smallButton("Refresh list", false).OnClick(p.refreshOpen)
	return p
}

func (p *modelsPage) open(providerID string) {
	p.providerID = providerID
	p.search.SetText("")
	if providerID != "" {
		p.search.Focus()
	}
}

func (p *modelsPage) refreshOpen() {
	if def, ok := catalog.Find(p.services.Settings, p.providerID); ok {
		p.services.Models.Refresh(def)
	}
}

func (p *modelsPage) Layout(gtx giode.Context) giode.Dimensions {
	if p.providerID == "" {
		return p.providerList().Layout(gtx)
	}
	return p.modelList().Layout(gtx)
}

func modelEntries(s *core.Services) []entry {
	m := &s.Settings.Models
	return []entry{
		{title: "Model for new chats", description: "Used when a chat starts; otherwise the most recent model is used.",
			keywords: "default model new chat", control: newModelChoice(s, &m.Default, "Most recent")},
		{title: "Model for chat titles", description: "A small, fast model is plenty. Empty uses the chat's own model.",
			keywords: "title naming model cheap fast", control: newModelChoice(s, &m.TitleModel, "Same as the chat")},
		{title: "Models in the picker", description: "Open a provider and untick the models you never use, to keep the picker short.",
			keywords: "hide show visible models picker list", wide: true, control: newModelsPage(s)},
	}
}
