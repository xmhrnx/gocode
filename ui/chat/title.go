package chat

import (
	"context"
	"strings"
	"time"

	"gocode/internal/catalog"
	"gocode/internal/llm"
	"gocode/internal/prompts"
)

const (
	maxTitleRunes = 60
	// titleTimeout bounds the naming call; the chat then keeps the
	// title made from the message.
	titleTimeout = 15 * time.Second
	// titleMaxTokens leaves room for a reasoning model to think first.
	titleMaxTokens = 4000
)

// maybeGenerateTitle names a chat that has a folder already: at once
// from the first message, then with a model call in the background.
func (c *ChatTab) maybeGenerateTitle(firstMessage string) {
	if c.session.Title != "" || strings.TrimSpace(firstMessage) == "" {
		return
	}
	c.session.Title = fallbackTitle(firstMessage)
	ask := c.titleRequest(firstMessage)
	if ask == nil {
		return
	}
	app := c.services.App
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), titleTimeout)
		defer cancel()
		if title := ask(ctx); title != "" {
			app.Post(func() { c.session.Title = title; c.save() })
		}
	}()
}

// titleRequest returns the call that asks the title model to name a
// chat, or nil when naming chats is off or the model can't be reached.
// The call returns "" when the model gives no title.
func (c *ChatTab) titleRequest(message string) func(context.Context) string {
	settings := c.services.Settings
	if !settings.General.AutoTitle || strings.TrimSpace(message) == "" {
		return nil
	}
	ref := catalog.ParseRef(settings.Models.TitleModel)
	if ref.IsZero() {
		ref = c.model
	}
	def, ok := catalog.Find(settings, ref.Provider)
	if !ok {
		return nil
	}
	system := prompts.Or(settings.Prompts.Title, prompts.DefaultTitle())
	keys, httpClient := c.services.Keys, c.services.HTTP
	return func(ctx context.Context) string {
		client, err := catalog.NewClient(def, keys, httpClient)
		if err != nil {
			return "" // the run will report the key problem
		}
		effort := ""
		if def.SendReasoningEffort {
			effort = "low" // reasoning models think less about a title
		}
		return requestTitle(ctx, client, llm.Request{Model: ref.Model, System: system, MaxTokens: titleMaxTokens,
			ReasoningEffort: effort, Messages: []llm.Message{{Role: llm.RoleUser,
				Parts: []llm.Part{llm.TextPart(truncateRunes(message, 2000))}}}})
	}
}

func requestTitle(ctx context.Context, client llm.Client, req llm.Request) string {
	var reply strings.Builder
	err := client.Stream(ctx, req, func(ev llm.Event) {
		if ev.Kind == llm.EventText {
			reply.WriteString(ev.Text)
		}
	})
	if err != nil {
		return ""
	}
	return cleanTitle(reply.String())
}

// cleanTitle keeps the title's first line, without quotes, a "Title:"
// label or a final full stop.
func cleanTitle(reply string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(reply), "\n")
	line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "Title:"))
	line = strings.Trim(line, "\"'`*. ")
	return truncateRunes(line, maxTitleRunes)
}

func fallbackTitle(message string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(message), "\n")
	return truncateRunes(line, 40)
}

func truncateRunes(s string, limit int) string {
	runes := []rune(s)
	if len(runes) <= limit {
		return s
	}
	return strings.TrimSpace(string(runes[:limit])) + "…"
}
