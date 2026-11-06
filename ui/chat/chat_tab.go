// Package chat is the agent chat screen: the transcript, the composer and
// the run that connects them to the model.
package chat

import (
	"os"

	"github.com/xmhrnx/giode"

	"gocode/internal/agent"
	"gocode/internal/catalog"
	"gocode/internal/llm"
	"gocode/internal/session"
	"gocode/internal/settings"
	"gocode/internal/tools"
	"gocode/ui/core"
)

// ChatTab is one conversation with the agent.
type ChatTab struct {
	services *core.Services
	session  *session.Session
	model    catalog.ModelRef
	perms    *agent.Permissions
	useTools bool
	registry *tools.Registry

	unsubscribe func()

	run *activeRun
	// queue holds messages sent while the agent works.
	queue   messageQueue
	todos   []tools.Todo
	notices []notice
	results map[string]*llm.ToolResult
	// contextTokens is the size of the last request, for the context gauge.
	contextTokens int
	revision      int

	header     *header
	transcript *transcript
	composer   *composer
}

// New returns a chat tab in workDir (empty: the default folder) with
// prompt in the composer.
func New(services *core.Services, workDir, prompt string) *ChatTab {
	cfg := services.Settings
	if workDir == "" {
		workDir = defaultWorkDir(cfg.General.DefaultWorkDir)
	}
	sess := session.New(workDir, cfg.Models.Default)
	return newTab(services, sess, prompt)
}

// Open returns a chat tab continuing a stored session.
func Open(services *core.Services, sess *session.Session) *ChatTab {
	return newTab(services, sess, "")
}

func newTab(services *core.Services, sess *session.Session, prompt string) *ChatTab {
	c := &ChatTab{
		services: services, session: sess, useTools: true,
		model:    initialModel(services, sess),
		perms:    agent.NewPermissions(agent.ModeAsk, services.Settings.Agent),
		registry: tools.Default(),
	}
	sess.Model = c.model.String()
	c.unsubscribe = services.OnSettingsChanged(func() {
		c.perms.SetPolicies(services.Settings.Agent)
		c.composer.applySettings()
		c.changed()
	})
	c.header = newHeader(c)
	c.transcript = newTranscript(c)
	c.composer = newComposer(c)
	c.composer.editor.SetText(prompt)
	c.composer.applySettings()
	return c
}

// View implements core.Tab.
func (c *ChatTab) View() giode.Element {
	giode.Shortcuts(
		giode.Shortcut("Mod+K", c.composer.picker.Open),
		giode.Shortcut("Esc", c.stop),
	)
	if c.services.Settings.Chat.SendKey == settings.SendCtrlEnter {
		giode.Shortcuts(giode.Shortcut("Mod+Enter", c.send))
	}
	return giode.Stack(giode.Styles{Width: -1, Height: -1, Background: "theme.background"},
		c.header,
		giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, c.transcript.element()),
		c.composer.view(),
	)
}

// changed marks the transcript as changed, so rows rebuild.
func (c *ChatTab) changed() { c.revision++ }

func initialModel(services *core.Services, sess *session.Session) catalog.ModelRef {
	if ref := catalog.ParseRef(sess.Model); !ref.IsZero() {
		return ref
	}
	if len(services.Settings.Models.Recent) > 0 {
		return catalog.ParseRef(services.Settings.Models.Recent[0])
	}
	return catalog.ModelRef{}
}

func defaultWorkDir(configured string) string {
	if configured != "" {
		return configured
	}
	if dir, err := os.Getwd(); err == nil && dir != "/" {
		return dir
	}
	if home, err := os.UserHomeDir(); err == nil {
		return home
	}
	return "."
}
