package settingsview

import (
	"fmt"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/checkbox"

	"gocode/internal/catalog"
	"gocode/internal/providers"
	"gocode/ui/theme"
)

// providerList shows each enabled provider with how many of its models
// the picker shows.
func (p *modelsPage) providerList() giode.Element {
	var rows []giode.Element
	for _, def := range catalog.Providers(p.services.Settings) {
		if catalog.Enabled(p.services.Settings, def.ID) {
			rows = append(rows, p.providerRow(def))
		}
	}
	if len(rows) == 0 {
		rows = append(rows, giode.Text("Enable a provider to choose its models.", theme.Caption()))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 2}, rows...)
}

func (p *modelsPage) providerRow(def providers.Definition) giode.Element {
	row, ok := p.providers[def.ID]
	if !ok {
		row = giode.Pressable(func(s giode.PressState) giode.Element {
			return navRowContent(def.Name, p.visibleCount(def), s)
		}).Label(def.Name).OnClick(func(giode.ClickEvent) { p.open(def.ID) })
		p.providers[def.ID] = row
	}
	return row
}

func (p *modelsPage) visibleCount(def providers.Definition) string {
	models := p.services.Models.Models(def)
	shown := 0
	for _, model := range models {
		if !p.services.Settings.Models.IsHidden(def.ID + "/" + model.ID) {
			shown++
		}
	}
	return fmt.Sprintf("%d of %d shown", shown, len(models))
}

// modelList shows the open provider's models matching the search.
func (p *modelsPage) modelList() giode.Element {
	def, ok := catalog.Find(p.services.Settings, p.providerID)
	if !ok {
		p.providerID = ""
		return p.providerList()
	}
	query := p.search.Text()
	var boxes []giode.Element
	for _, model := range p.services.Models.Models(def) {
		if (entry{title: model.ID + " " + model.Name}).matches(query) {
			boxes = append(boxes, giode.Box(giode.Styles{Width: 290}, p.box(def.ID+"/"+model.ID, model.ID)))
		}
	}
	if len(boxes) == 0 {
		boxes = append(boxes, giode.Text("No models match.", theme.Caption()))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 10},
		giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			p.back, giode.Box(giode.Styles{FlexGrow: 1}, p.search), p.refresh),
		groupTitle(def.Name),
		giode.Row(giode.Styles{Width: -1, Gap: 4}, boxes...),
	)
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

// navRowContent draws a row that opens a sub-page.
func navRowContent(name, detail string, s giode.PressState) giode.Element {
	background := ""
	if s.Hovered {
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
