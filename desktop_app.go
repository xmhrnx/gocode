//go:build !android && !ios

package main

import (
	"image"
	"path/filepath"
	"strings"
	"time"

	"github.com/xmhrnx/giode"

	"gocode/assets"
	"gocode/internal/browser"
	"gocode/internal/catalog"
	"gocode/internal/mcp"
	"gocode/internal/remote/host"
	"gocode/internal/secrets"
	"gocode/internal/session"
	"gocode/internal/settings"
	"gocode/ui"
	"gocode/ui/core"
	"gocode/ui/phone"
	"gocode/ui/platform"
)

// run starts the desktop app: the agent, its tabs and settings.
func run() error {
	tuneMemory()
	if err := assets.Load(); err != nil {
		return err
	}
	dir, err := settings.Dir()
	if err != nil {
		return err
	}
	services, settingsErr := newServices(dir)
	var root *ui.Root
	app := giode.New(giode.Config{
		Title: "GoCode", Size: image.Pt(1240, 820), MinWidth: 560, MinHeight: 420,
		Resizable: true, Frameless: true, Center: true, Background: "theme.background",
		Icon:       assets.AppIcon,
		OnFileDrop: func(paths []string) { root.DropFiles(paths) },
	})
	services.App = app
	services.Toasts = giode.Toaster().Placement(giode.ToastBottomEnd).Max(3).Inset(giode.Inset{Bottom: 120, Right: 16})
	services.Models = catalog.NewModelStore(services.Keys, services.HTTP, func() { app.Post(func() {}) })
	services.MCP.OnChange(func() { app.Post(func() {}) })
	services.MCP.Sync(services.Settings.MCPServers)
	services.Browser.OnChange(func() { app.Post(func() {}) })
	syncBrowser := func() { services.Browser.Sync(services.Settings.Browser.Enabled, services.Settings.Browser.Port) }
	syncBrowser()
	services.OnSettingsChanged(func() {
		services.MCP.Sync(services.Settings.MCPServers)
		syncBrowser()
	})
	defer services.MCP.Close()
	defer services.Browser.Close()
	phoneErr := startPhoneHost(services, dir, func() []core.Tab { return root.Tabs() })
	if services.Phone != nil {
		defer services.Phone.Close()
	}
	root = ui.New(services)
	if phoneErr != nil {
		services.Error("The phone remote is off: " + phoneErr.Error())
	}
	if settingsErr != nil {
		services.Error(settingsErr.Error() + " Using defaults.")
	}
	if err := app.Run(root.View); err != nil {
		return err
	}
	return root.SaveTabsBeforeExit()
}

// startPhoneHost sets up the phone remote: the desktop's identity, the
// paired phones, and listening as the settings say.
func startPhoneHost(services *core.Services, dir string, tabs func() []core.Tab) error {
	key, room, err := phone.LoadIdentity(services.Keys)
	if err != nil {
		return err
	}
	devices, err := host.OpenDevices(filepath.Join(dir, "phones.json"))
	if err != nil {
		return err
	}
	name := phone.DesktopName()
	services.Phone = host.New(host.Config{Name: name, Key: key, Devices: devices, Room: room,
		Controller: phone.NewController(services.App, name, tabs)})
	services.Phone.OnChange(func() { services.App.Post(func() {}) })
	sync := func() {
		p := services.Settings.Phone
		services.Phone.Sync(p.Enabled, p.Port, strings.TrimSpace(p.Relay))
	}
	sync()
	services.OnSettingsChanged(sync)
	return nil
}

// newServices loads the settings and opens the stores. A broken settings
// file is reported, not fatal: the app starts with defaults.
func newServices(dir string) (*core.Services, error) {
	loaded, err := settings.Load(filepath.Join(dir, "settings.json"))
	timeout := time.Duration(loaded.Network.RequestTimeoutSeconds) * time.Second
	return &core.Services{
		Settings:     &loaded,
		SettingsPath: filepath.Join(dir, "settings.json"),
		Keys: secrets.Fallback{
			Primary:   platform.NewKeychain(),
			Secondary: &secrets.FileStore{Path: filepath.Join(dir, "keys.json")},
		},
		Sessions:  session.Store{Dir: filepath.Join(dir, "sessions")},
		HTTP:      platform.NewHTTPClient(timeout),
		MCP:       mcp.NewManager(),
		Browser:   browser.New(),
		SkillsDir: filepath.Join(dir, "skills"),
		TabsPath:  filepath.Join(dir, "tabs.json"),
	}, err
}
