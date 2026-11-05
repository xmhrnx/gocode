// Command gocode is a native, lightweight coding agent and terminal.
package main

import (
	"image"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/xmhrnx/giode"

	"gocode/assets"
	"gocode/internal/catalog"
	"gocode/internal/mcp"
	"gocode/internal/secrets"
	"gocode/internal/session"
	"gocode/internal/settings"
	"gocode/ui"
	"gocode/ui/core"
	"gocode/ui/platform"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

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
	services.OnSettingsChanged(func() { services.MCP.Sync(services.Settings.MCPServers) })
	defer services.MCP.Close()
	root = ui.New(services)
	if settingsErr != nil {
		services.Error(settingsErr.Error() + " Using defaults.")
	}
	return app.Run(root.View)
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
		SkillsDir: filepath.Join(dir, "skills"),
	}, err
}

// tuneMemory makes the garbage collector run when the heap grows by half
// instead of doubling. A streaming chat makes a lot of short-lived garbage
// (each frame lays out new text), and at Go's default the heap holds up to
// twice what is live; the extra collections cost little CPU. GOGC, when
// set, still wins.
func tuneMemory() {
	if os.Getenv("GOGC") == "" {
		debug.SetGCPercent(50)
	}
}
