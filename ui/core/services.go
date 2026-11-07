// Package core holds what every screen shares: the app, the settings, the
// stores, and the Tab contract the workspace manages.
package core

import (
	"net/http"
	"sync"
	"time"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/toaster"

	"gocode/internal/catalog"
	"gocode/internal/mcp"
	"gocode/internal/secrets"
	"gocode/internal/session"
	"gocode/internal/settings"
)

// Services is shared by every screen. Its fields are read and changed on
// the UI goroutine unless noted.
type Services struct {
	App          *giode.App
	Settings     *settings.Settings
	SettingsPath string
	// Keys, Sessions, Models and HTTP are safe from any goroutine.
	Keys     secrets.Store
	Sessions session.Store
	Models   *catalog.ModelStore
	HTTP     *http.Client
	// MCP connects to the user's MCP servers; safe from any goroutine.
	MCP *mcp.Manager
	// SkillsDir is where installed skills live.
	SkillsDir string
	// TabsPath is where the open tabs are remembered.
	TabsPath string
	Toasts   *toaster.Toaster

	// soundTrouble reports, once, that sounds can't play.
	soundTrouble sync.Once

	saveTimer   *time.Timer
	listeners   map[int]func()
	nextListen  int
	revision    int
	openFactory OpenFactory
}

// OpenFactory opens new tabs; the workspace provides it.
type OpenFactory interface {
	// OpenChat opens a chat in workDir (empty: the default folder) with
	// prompt in its composer.
	OpenChat(workDir, prompt string)
	OpenSettings(category string)
}

// SetOpenFactory connects the workspace.
func (s *Services) SetOpenFactory(f OpenFactory) { s.openFactory = f }

// Open returns the tab opener.
func (s *Services) Open() OpenFactory { return s.openFactory }

// SettingsRevision changes every time the settings change, for memo deps.
func (s *Services) SettingsRevision() int { return s.revision }

// OnSettingsChanged registers fn to run after every settings change and
// returns a function that unregisters it, for tabs that close.
func (s *Services) OnSettingsChanged(fn func()) (unsubscribe func()) {
	if s.listeners == nil {
		s.listeners = map[int]func(){}
	}
	s.nextListen++
	id := s.nextListen
	s.listeners[id] = fn
	return func() { delete(s.listeners, id) }
}

// SettingsChanged notifies listeners and saves the settings shortly after,
// coalescing bursts of changes (dragging a slider) into one write.
func (s *Services) SettingsChanged() {
	s.revision++
	for _, fn := range s.listeners {
		fn()
	}
	if s.saveTimer != nil {
		s.saveTimer.Stop()
	}
	snapshot := *s.Settings
	s.saveTimer = time.AfterFunc(400*time.Millisecond, func() {
		if err := settings.Save(s.SettingsPath, snapshot); err != nil {
			s.Error("Couldn't save settings: " + err.Error())
		}
	})
}

// Error shows an error toast; safe from any goroutine.
func (s *Services) Error(message string) {
	s.Toasts.Show(giode.ToastError, message)
}

// Info shows an informational toast; safe from any goroutine.
func (s *Services) Info(message string) {
	s.Toasts.Show(giode.ToastInfo, message)
}
