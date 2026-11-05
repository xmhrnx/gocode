package settingsview

import (
	"gioui.org/layout"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/toggle"

	"gocode/internal/catalog"
	"gocode/internal/providers"
	"gocode/ui/core"
	"gocode/ui/theme"
)

// providerRow is one provider in the list: an on/off switch, its name and
// key status; clicking it opens its editor.
type providerRow struct {
	services *core.Services
	page     *providersPage
	id       string
	enabled  *toggle.Switch
	header   *giode.PressableElement
	editor   *providerEditor
	expanded bool
}

func newProviderRow(page *providersPage, def providers.Definition) *providerRow {
	s := page.services
	r := &providerRow{services: s, page: page, id: def.ID}
	r.enabled = giode.Input.Switch("")
	r.enabled.SetChecked(catalog.Enabled(s.Settings, def.ID))
	r.enabled.OnChange(func(on bool) {
		providerSettings(s, def.ID).Enabled = on
		s.SettingsChanged()
		if on {
			if current, ok := catalog.Find(s.Settings, def.ID); ok {
				s.Models.Refresh(current)
			}
		}
	})
	r.header = giode.Pressable(r.headerContent).Label(def.Name).OnClick(func(giode.ClickEvent) { r.expanded = !r.expanded })
	r.editor = newProviderEditor(page, def)
	return r
}

func (r *providerRow) definition() providers.Definition {
	def, _ := catalog.Find(r.services.Settings, r.id)
	return def
}

func (r *providerRow) Layout(gtx layout.Context) layout.Dimensions {
	children := []giode.Element{
		giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Gap: 12, Wrap: giode.WrapNoWrap, Padding: giode.Inset{Left: 12}},
			r.enabled, giode.Box(giode.Styles{FlexGrow: 1}, r.header)),
	}
	if r.expanded {
		children = append(children, r.editor)
	}
	background := ""
	if r.expanded {
		background = theme.Raised
	}
	return giode.Stack(giode.Styles{Width: -1, BorderRadius: 10, Background: background}, children...).Layout(gtx)
}

func (r *providerRow) headerContent(state giode.PressState) giode.Element {
	def := r.definition()
	color := theme.Muted
	if state.Hovered {
		color = theme.Text
	}
	chevron := "chevron-right"
	if r.expanded {
		chevron = "chevron-down"
	}
	return giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Gap: 10, Wrap: giode.WrapNoWrap, Padding: giode.SymmetricInset(4, 10)},
		giode.Stack(giode.Styles{FlexGrow: 1, Gap: 2},
			giode.Text(def.Name, giode.Styles{Color: theme.Text, FontWeight: giode.FontWeightMedium}),
			giode.Text(def.BaseURL, giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Tiny, FontFamily: "JetBrains Mono",
				TextOverflow: giode.TextOverflowEllipsis}),
		),
		r.page.keyBadge(def),
		theme.Icon(chevron, 14, color),
	)
}
