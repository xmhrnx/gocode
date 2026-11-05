package modelpicker

import (
	"github.com/xmhrnx/giode"

	"gocode/internal/catalog"
	"gocode/ui/theme"
)

// modelsView lists the open provider's visible models matching the search.
func (p *Picker) modelsView() giode.Element {
	def, ok := catalog.Find(p.services.Settings, p.providerID)
	if !ok {
		return message("This provider is no longer configured.")
	}
	query := p.search.Text()
	p.firstMatch = nil
	status := p.services.Models.Status(def.ID)
	var rows []giode.Element
	if status.Err != nil {
		rows = append(rows, message("Couldn't list models: "+status.Err.Error()))
	}
	for _, model := range p.services.Models.Models(def) {
		ref := def.ID + "/" + model.ID
		if p.services.Settings.Models.IsHidden(ref) || !matches(query, model.ID+" "+model.Name) {
			continue
		}
		if p.firstMatch == nil {
			p.firstMatch = func() { p.choose(ref) }
		}
		rows = append(rows, p.modelRow(ref, ref, model.ID, ""))
	}
	if len(rows) == 0 {
		rows = append(rows, message(emptyModelsText(query, status.Loading)))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 1, Padding: giode.Inset{Right: 8}}, rows...)
}

// modelRow is a choosable model; detail (the provider, for recent models)
// shows on the right.
func (p *Picker) modelRow(key, ref, model, detail string) giode.Element {
	return p.modelRows.Get(key, func() *giode.PressableElement {
		return giode.Pressable(func(s giode.PressState) giode.Element {
			return modelRowContent(model, detail, s, ref == p.current())
		}).Label(model).OnClick(func(giode.ClickEvent) { p.choose(ref) })
	})
}

func modelRowContent(model, detail string, s giode.PressState, selected bool) giode.Element {
	background := ""
	if s.Hovered || s.Focused {
		background = theme.Hover
	}
	check := giode.Element(giode.Box(giode.Styles{Width: 14}))
	if selected {
		check = theme.Icon("check", 14, theme.Primary)
	}
	return giode.Row(giode.Styles{Width: -1, Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
		Padding: giode.SymmetricInset(10, 7), BorderRadius: 8, Background: background},
		check,
		giode.Text(model, giode.Styles{Color: theme.Text, FlexGrow: 1, TextOverflow: giode.TextOverflowEllipsis}),
		giode.Text(detail, giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Small}),
	)
}

func emptyModelsText(query string, loading bool) string {
	switch {
	case loading:
		return "Loading models…"
	case query != "":
		return "No models match “" + query + "”."
	}
	return "This provider lists no models. Add model IDs in Settings → Providers."
}
