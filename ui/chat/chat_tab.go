// Package chat is the agent chat screen: the transcript, the composer and
// the run that connects them to the model.
package chat

import (
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
	// effort is this chat's reasoning effort; it starts at the default
	// from the settings.
	effort   string
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
	// toggledReasoning holds the reasoning parts the user opened or
	// folded against the setting's default, by their item group.
	toggledReasoning map[string]bool

	header     *header
	transcript *transcript
	composer   *composer
}

// New returns a chat tab in workDir with prompt in the composer. With no
// workDir, the chat gets a folder of its own in the projects folder when
// its first message is sent, unless the user opens one before.
func New(services *core.Services, workDir, prompt string) *ChatTab {
	sess := session.New(workDir, services.Settings.Models.Default)
	return newTab(services, sess, prompt)
}

// Open returns a chat tab continuing a stored session.
func Open(services *core.Services, sess *session.Session) *ChatTab {
	return newTab(services, sess, "")
}

func newTab(services *core.Services, sess *session.Session, prompt string) *ChatTab {
	c := &ChatTab{
		services: services, session: sess, useTools: true, effort: services.Settings.Agent.ReasoningEffort,
		model:    initialModel(services, sess),
		perms:    agent.NewPermissions(agent.ModeAsk, services.Settings.Agent),
		registry: tools.Default(), toggledReasoning: map[string]bool{},
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
