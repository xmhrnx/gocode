package chat

import (
	"context"
	"errors"
	"time"

	"gocode/internal/agent"
	"gocode/internal/catalog"
	"gocode/internal/chime"
	"gocode/internal/llm"
	"gocode/internal/tools"
)

// activeRun is the agent run in progress, and what it has streamed so far.
type activeRun struct {
	cancel  context.CancelFunc
	started time.Time
	draft   llm.PartBuilder
	running map[string]bool // tool call IDs currently running
	// finished holds results of this step's calls until the agent adds
	// them to the conversation together.
	finished  map[string]llm.ToolResult
	callUsage llm.Usage
	// outputTokens counts what the run's finished model calls wrote, and
	// streamedBytes the text and reasoning of the current call so far.
	outputTokens  int
	streamedBytes int
	status        string
	ask           *pendingPermission
	// subagents holds the progress of running tasks, by task call ID.
	subagents map[string]*subagentProgress
}

// send adds the composer's content as a user message and starts a run.
func (c *ChatTab) send() {
	msg, ok := c.composer.takeMessage()
	if !ok {
		return
	}
	c.composer.editor.Focus()
	if c.run != nil {
		c.queue.push(msg) // delivered at the agent's next step
		c.changed()
		return
	}
	if c.model.IsZero() {
		c.composer.restore(msg)
		c.services.Error("Pick a model first: click the model button below the message box.")
		return
	}
	c.session.Messages = append(c.session.Messages, msg)
	if c.session.WorkDir == "" {
		c.nameThenRun(msg)
		return
	}
	c.maybeGenerateTitle(msg.Text())
	c.startRun()
}

// startRun runs the agent on the conversation so far.
func (c *ChatTab) startRun() {
	def, ok := catalog.Find(c.services.Settings, c.model.Provider)
	if !ok {
		c.addNotice(noticeError, "The provider "+c.model.Provider+" isn't configured. Open Settings → Providers.")
		return
	}
	cfg := c.agentConfig()
	history := c.session.ModelHistory(agent.SummaryMessage)
	ctx, cancel := context.WithCancel(context.Background())
	c.run = newActiveRun(cancel)
	c.services.Settings.Models.MarkUsed(c.model.String())
	c.services.SettingsChanged()
	app, keys, httpClient, vars := c.services.App, c.services.Keys, c.services.HTTP, c.promptVars()
	go func() {
		client, err := catalog.NewClient(def, keys, httpClient)
		if err == nil {
			cfg.Client = client
			cfg.System = c.systemPrompt(vars)
			_, err = agent.Run(ctx, cfg, history, func(ev agent.Event) {
				app.Post(func() { c.handleEvent(ev) })
			})
		}
		app.Post(func() { c.finishRun(err) })
	}()
	c.changed()
}

// stop cancels the run in progress.
// tokens is how many tokens the run has written: the usage the provider
// reported, or while a call streams, an estimate from its text.
func (r *activeRun) tokens() int {
	return r.outputTokens + r.callTokens()
}

// callTokens is the current call's output: as reported, or about four
// bytes of text per token until it is.
func (r *activeRun) callTokens() int {
	return max(r.callUsage.OutputTokens, r.streamedBytes/4)
}

func (c *ChatTab) stop() {
	if c.run != nil {
		c.run.cancel()
	}
}

func (c *ChatTab) finishRun(err error) {
	c.run = nil
	if queued := c.queue.drain(); len(queued) > 0 {
		if err == nil {
			// Sent just as the run ended: answer them in a new run.
			c.session.Messages = append(c.session.Messages, queued...)
			c.save()
			c.startRun()
			return
		}
		c.composer.restoreQueued(queued) // the run stopped; let the user decide
	}
	// The plan shows the run's progress; a model that didn't tick its
	// last items off would otherwise leave a stale plan on screen. The
	// list stays in the transcript, in its todo_write calls.
	c.todos = nil
	switch {
	case err == nil:
		c.services.PlaySound(chime.Done)
	case errors.Is(err, context.Canceled):
		c.addNotice(noticeInfo, "Stopped.")
	default:
		c.addNotice(noticeError, err.Error())
		c.services.PlaySound(chime.Failure)
	}
	c.save()
	c.changed()
}

// setTodos receives the agent's task list.
func (c *ChatTab) setTodos(todos []tools.Todo) {
	c.todos = todos
	c.changed()
}

// draftMessage returns the reply streamed so far.
func (r *activeRun) draftMessage() llm.Message {
	return llm.Message{Role: llm.RoleAssistant, Parts: r.draft.Parts()}
}
