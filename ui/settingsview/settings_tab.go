// Package settingsview is the settings tab: categories on the left, a
// search field that finds any setting, and the settings on the right.
package settingsview

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"
	"github.com/xmhrnx/giode/components/scroll"

	"gocode/internal/tabstate"
	"gocode/ui/core"
	"gocode/ui/theme"

	"gocode/ui/components"
)

// Tab is the settings page.
type Tab struct {
	services *core.Services
	search   *input.Input
	active   string
	entries  []entry
	nav      []giode.Element
	content  *scroll.Scroll
}

// New returns the settings tab showing category, or the first one.
func New(services *core.Services, category string) *Tab {
	t := &Tab{services: services, active: categories[0].name}
	t.search = giode.Input.Text("Search settings…", giode.Styles{Background: theme.Canvas, BorderColor: theme.Border,
		BorderRadius: 9, Color: theme.Text, FontSize: theme.Sizes.Small})
	components.AddEditMenu(services.App, t.search, components.EditOptions{})
	t.entries = allEntries(services)
	for _, c := range categories {
		t.nav = append(t.nav, newNavItem(t, c))
	}
	t.content = giode.Scroll(t.page, giode.Styles{ScrollBar: giode.ScrollBar{Width: 7, AutoHide: true, TrackColor: "transparent", Inset: 3}})
	t.ShowCategory(category)
	return t
}

func allEntries(s *core.Services) []entry {
	builders := map[string]func() []entry{
		"General": func() []entry { return generalEntries(s) }, "Appearance": func() []entry { return appearanceEntries(s) },
		"Chat": func() []entry { return chatEntries(s) }, "Agent": func() []entry { return agentEntries(s) },
		"Permissions": func() []entry { return permissionEntries(s) }, "Prompts": func() []entry { return promptEntries(s) },
		"Providers": func() []entry { return providerEntries(s) }, "Models": func() []entry { return modelEntries(s) },
		"Skills": func() []entry { return skillEntries(s) }, "MCP servers": func() []entry { return mcpEntries(s) },
		"Web search": func() []entry { return webSearchEntries(s) },
		"Sounds":     func() []entry { return soundEntries(s) },
		"Network":    func() []entry { return networkEntries(s) },
		"Shortcuts":  shortcutEntries, "About": func() []entry { return aboutEntries(s) },
	}
	var all []entry
	for _, c := range categories {
		for _, e := range builders[c.name]() {
			e.category = c.name
			all = append(all, e)
		}
	}
	return all
}

// ShowCategory switches to the named category.
func (t *Tab) ShowCategory(name string) {
	for _, c := range categories {
		if c.name == name {
			t.active = name
			t.search.SetText("")
			t.content.ScrollTo(0)
		}
	}
}

// Title implements core.Tab.
func (t *Tab) Title() string { return "Settings" }

// Snapshot implements core.Restorable.
func (t *Tab) Snapshot() tabstate.Tab { return tabstate.Tab{Kind: tabstate.KindSettings} }

// Icon implements core.Tab.
func (t *Tab) Icon() string { return "settings" }

// Busy implements core.Tab.
func (t *Tab) Busy() bool { return false }

// Close implements core.Tab.
func (t *Tab) Close() {}

// Focus implements core.Focusable.
func (t *Tab) Focus() { t.search.Focus() }

// View implements core.Tab.
func (t *Tab) View() giode.Element {
	return giode.Row(giode.Styles{Width: -1, Height: -1, Wrap: giode.WrapNoWrap, Align: giode.AlignStretch},
		giode.Stack(giode.Styles{Width: 236, Gap: 2, Padding: giode.UniformInset(14), Background: theme.Panel},
			giode.Box(giode.Styles{Width: -1, Padding: giode.Inset{Bottom: 10}}, t.search),
			giode.Stack(giode.Styles{Width: -1, Gap: 2}, t.nav...),
		),
		giode.Box(giode.Styles{Width: 1, Background: theme.SoftBorder}),
		giode.Box(giode.Styles{FlexGrow: 1, Height: -1}, t.content),
	)
}
