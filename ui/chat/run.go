package chat

import (
	"context"
	"errors"
	"time"

	"gocode/internal/agent"
	"gocode/internal/catalog"
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
	status    string
	ask       *pendingPermission
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
	c.run = &activeRun{cancel: cancel, started: time.Now(), running: map[string]bool{}, finished: map[string]llm.ToolResult{}}
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
	switch {
	case err == nil:
	case errors.Is(err, context.Canceled):
		c.addNotice(noticeInfo, "Stopped.")
	default:
		c.addNotice(noticeError, err.Error())
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
