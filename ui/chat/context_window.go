package chat

import (
	"context"
	"time"

	"gocode/internal/agent"
	"gocode/internal/catalog"
)

// contextWindow returns the chat model's context size: what the provider
// reports for it, or the size from the settings when it reports nothing.
func (c *ChatTab) contextWindow() int {
	if def, ok := catalog.Find(c.services.Settings, c.model.Provider); ok {
		for _, model := range c.services.Models.Models(def) {
			if model.ID == c.model.Model && model.ContextWindow > 0 {
				return model.ContextWindow
			}
		}
	}
	return c.services.Settings.Agent.ContextWindow
}

// compactNow summarizes the conversation on request, so a long chat can
// continue with a fresh context.
func (c *ChatTab) compactNow() {
	def, ok := catalog.Find(c.services.Settings, c.model.Provider)
	if c.run != nil || !ok || len(c.session.Messages) == 0 {
		return
	}
	cfg := c.agentConfig()
	history := c.session.ModelHistory(agent.SummaryMessage)
	ctx, cancel := context.WithCancel(context.Background())
	c.run = &activeRun{cancel: cancel, started: time.Now(), status: "Compacting the conversation"}
	app, keys, httpClient := c.services.App, c.services.Keys, c.services.HTTP
	go func() {
		client, err := catalog.NewClient(def, keys, httpClient)
		summary := ""
		if err == nil {
			cfg.Client = client
			summary, err = agent.Compact(ctx, cfg, history)
		}
		app.Post(func() {
			if err == nil {
				c.handleEvent(agent.Event{Kind: agent.EventCompacted, Text: summary})
			}
			c.finishRun(err)
		})
	}()
	c.changed()
}
