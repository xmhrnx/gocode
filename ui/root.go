// Package ui builds GoCode's window: the title bar with the tab strip,
// the active tab, and the overlays shared by every tab.
package ui

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/titlebar"

	"gocode/internal/catalog"
	"gocode/internal/session"
	"gocode/ui/chat"
	"gocode/ui/components"
	"gocode/ui/core"
	"gocode/ui/history"
	"gocode/ui/settingsview"
	"gocode/ui/theme"
	"gocode/ui/workspace"
)

// Root is the whole window.
type Root struct {
	services  *core.Services
	workspace *workspace.Workspace
	bar       *titlebar.TitleBar
	history   *history.Dialog
	closing   *closeConfirmation
}

// New builds the window for services and opens the first tab.
func New(services *core.Services) *Root {
	r := &Root{services: services}
	r.workspace = workspace.New(workspace.Factory{
		Chat:     func(workDir, prompt string) core.Tab { return chat.New(services, workDir, prompt) },
		Settings: func(category string) core.Tab { return settingsview.New(services, category) },
	})
	services.SetOpenFactory(r.workspace)
	r.history = history.New(services, func(sess *session.Session) { r.workspace.Open(chat.Open(services, sess)) })
	r.closing = newCloseConfirmation(services.App)
	r.bar = r.newTitleBar()
	services.OnSettingsChanged(r.applyAppearance)
	r.applyAppearance()
	if !r.restoreTabs() {
		r.workspace.OpenChat("", "")
	}
	r.workspace.OnChange(r.SaveTabs)
	r.refreshModelLists()
	return r
}

// refreshModelLists fetches the enabled providers' models in the
// background, so the model picker is ready when first opened.
func (r *Root) refreshModelLists() {
	for _, def := range catalog.Providers(r.services.Settings) {
		if catalog.Enabled(r.services.Settings, def.ID) {
			r.services.Models.Refresh(def)
		}
	}
}

func (r *Root) newTitleBar() *titlebar.TitleBar {
	historyButton := components.NewIconButton("history", "Chat history  ("+giode.FormatShortcut("Mod+H")+")", r.history.Open)
	settingsButton := components.NewIconButton("settings", "Settings  ("+giode.FormatShortcut("Mod+,")+")", func() { r.workspace.OpenSettings("") })
	return giode.TitleBar(r.services.App, "", giode.Styles{
		Background: theme.Panel, Color: theme.Muted, Height: 40, BorderWidth: 1, BorderColor: theme.SoftBorder,
	}).
		Icon(giode.Box(giode.Styles{Padding: giode.Inset{Left: 4}}, theme.Icon("square-terminal", 18, theme.Primary))).
		Leading(workspace.NewTabStrip(r.workspace)).
		Trailing(historyButton, settingsButton).
		OnClose(func() bool {
			return r.closing.allow(r.workspace.AnyBusy() && r.services.Settings.General.ConfirmCloseBusy)
		})
}

// applyAppearance applies the theme, text sizes and title bar look.
func (r *Root) applyAppearance() {
	appearance := r.services.Settings.Appearance
	theme.Apply(appearance.Theme)
	theme.SetBaseSize(appearance.UIFontSize, appearance.CodeFontSize)
	r.services.App.SetMaxFPS(max(appearance.MaxFPS, 0)) // NoFrameCap (-1) is Giode's 0: no cap
	if appearance.TitleBarStyle == "mac" {
		r.bar.Variant(giode.TitleBarMac)
	} else {
		r.bar.Variant(giode.TitleBarWindows)
	}
	giode.InvalidateCaches()
}

// DropFiles hands files dropped on the window to the active tab.
func (r *Root) DropFiles(paths []string) {
	if !r.workspace.DropFiles(paths) {
		r.services.Info("Open a chat to attach files.")
	}
}

// View builds the window's content; it runs every frame.
func (r *Root) View() giode.Element {
	r.registerShortcuts()
	content := giode.Element(giode.Box(giode.Styles{}))
	if tab := r.workspace.Active(); tab != nil {
		content = tab.View()
	}
	return giode.Stack(giode.Styles{Width: -1, Height: -1, Background: theme.Canvas},
		r.bar,
		giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, content),
		r.history,
		r.closing,
		r.services.Toasts,
	)
}
