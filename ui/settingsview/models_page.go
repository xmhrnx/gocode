package settingsview

import (
	"gioui.org/layout"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/checkbox"
	"github.com/xmhrnx/giode/components/input"

	"gocode/internal/catalog"
	"gocode/ui/core"
	"gocode/ui/theme"
)

// modelsPage lists every enabled provider's models with a checkbox that
// shows or hides each in the model picker.
type modelsPage struct {
	services *core.Services
	search   *input.Input
	boxes    *giode.KeyedCache[string, *checkbox.Checkbox]
	refresh  *giode.ButtonElement
}

func modelEntries(s *core.Services) []entry {
	m := &s.Settings.Models
	return []entry{
		{title: "Model for new chats", description: "Used when a chat starts; otherwise the most recent model is used.",
			keywords: "default model new chat", control: newModelChoice(s, &m.Default, "Most recent")},
		{title: "Model for chat titles", description: "A small, fast model is plenty. Empty uses the chat's own model.",
			keywords: "title naming model cheap fast", control: newModelChoice(s, &m.TitleModel, "Same as the chat")},
		{title: "Models in the picker", description: "Untick models you never use to keep the picker short.",
			keywords: "hide show visible models picker list", wide: true, control: newModelsPage(s)},
	}
}

func newModelsPage(s *core.Services) *modelsPage {
	p := &modelsPage{services: s, boxes: giode.NewKeyedCache[string, *checkbox.Checkbox]()}
	p.search = giode.Input.Text("Filter models…", fieldStyle(-1))
	p.refresh = smallButton("Refresh lists", false).OnClick(p.refreshAll)
	return p
}

func (p *modelsPage) refreshAll() {
	for _, def := range catalog.Providers(p.services.Settings) {
		if catalog.Enabled(p.services.Settings, def.ID) {
			p.services.Models.Refresh(def)
		}
	}
}

func (p *modelsPage) Layout(gtx layout.Context) layout.Dimensions {
	query := p.search.Text()
	children := []giode.Element{giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
		giode.Box(giode.Styles{FlexGrow: 1}, p.search), p.refresh)}
	for _, def := range catalog.Providers(p.services.Settings) {
		if !catalog.Enabled(p.services.Settings, def.ID) {
			continue
		}
		var boxes []giode.Element
		for _, model := range p.services.Models.Models(def) {
			ref := def.ID + "/" + model.ID
			if (entry{title: def.Name + " " + model.ID}).matches(query) {
				boxes = append(boxes, giode.Box(giode.Styles{Width: 290}, p.box(ref, model.ID)))
			}
		}
		if len(boxes) > 0 {
			children = append(children, groupTitle(def.Name), giode.Row(giode.Styles{Width: -1, Gap: 4}, boxes...))
		}
	}
	if len(children) == 1 {
		children = append(children, giode.Text("Enable a provider to manage its models.", theme.Caption()))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 8}, children...).Layout(gtx)
}

func (p *modelsPage) box(ref, label string) giode.Element {
	models := &p.services.Settings.Models
	return p.boxes.Get(ref, func() *checkbox.Checkbox {
		box := giode.Input.Checkbox(label, giode.Styles{Color: theme.Text, FontSize: theme.Sizes.Small})
		box.SetChecked(!models.IsHidden(ref))
		box.OnChange(func(visible bool) {
			models.SetHidden(ref, !visible)
			p.services.SettingsChanged()
		})
		return box
	})
}
