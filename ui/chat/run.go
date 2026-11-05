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
}

// send adds the composer's content as a user message and starts a run.
func (c *ChatTab) send() {
	if c.run != nil {
		return
	}
	msg, ok := c.composer.takeMessage()
	if !ok {
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

func (c *ChatTab) handleEvent(ev agent.Event) {
	run := c.run
	if run == nil {
		return
	}
	switch ev.Kind {
	case agent.EventText:
		run.draft.AppendText(llm.PartText, ev.Text)
	case agent.EventReasoning:
		run.draft.AppendText(llm.PartReasoning, ev.Text)
	case agent.EventToolStarted:
		run.running[ev.Call.ID] = true
	case agent.EventToolFinished:
		delete(run.running, ev.Call.ID)
		run.finished[ev.Call.ID] = llm.ToolResult{CallID: ev.Call.ID, Name: ev.Call.Name,
			Content: ev.Result.Output, IsError: ev.Result.IsError}
		if ev.Result.Display != "" {
			c.session.ToolDisplays[ev.Call.ID] = ev.Result.Display
		}
	case agent.EventMessage:
		c.commitMessage(*ev.Message)
	case agent.EventUsage:
		run.callUsage = *ev.Usage
		c.contextTokens = ev.Usage.InputTokens + ev.Usage.CacheReadTokens + ev.Usage.CacheWriteTokens
	case agent.EventRetry:
		run.status = ev.Text
	case agent.EventNotice:
		c.addNotice(noticeInfo, ev.Text)
	case agent.EventCompacted:
		c.session.Compacted(ev.Text)
		c.contextTokens = 0
		c.addNotice(noticeInfo, "Context compacted: earlier messages were summarized for the model. They stay visible here.")
	}
	if ev.Kind != agent.EventRetry {
		run.status = ""
	}
	c.changed()
}

func (c *ChatTab) commitMessage(msg llm.Message) {
	c.session.Messages = append(c.session.Messages, msg)
	if msg.Role == llm.RoleAssistant {
		c.session.Usage = c.session.Usage.Add(c.run.callUsage)
		c.run.callUsage = llm.Usage{}
		c.run.draft = llm.PartBuilder{}
	}
	c.save()
}

func (c *ChatTab) finishRun(err error) {
	c.run = nil
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
