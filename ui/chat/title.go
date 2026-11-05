package chat

import (
	"context"
	"strings"
	"time"

	"gocode/internal/catalog"
	"gocode/internal/llm"
	"gocode/internal/prompts"
)

const maxTitleRunes = 60

// maybeGenerateTitle names the chat after its first message: at once from
// the text, then, when enabled, with a short model call in the background.
func (c *ChatTab) maybeGenerateTitle(firstMessage string) {
	if c.session.Title != "" || strings.TrimSpace(firstMessage) == "" {
		return
	}
	c.session.Title = fallbackTitle(firstMessage)
	settings := c.services.Settings
	if !settings.General.AutoTitle {
		return
	}
	ref := catalog.ParseRef(settings.Models.TitleModel)
	if ref.IsZero() {
		ref = c.model
	}
	def, ok := catalog.Find(settings, ref.Provider)
	if !ok {
		return
	}
	system := prompts.Or(settings.Prompts.Title, prompts.DefaultTitle())
	app, keys, httpClient := c.services.App, c.services.Keys, c.services.HTTP
	go func() {
		client, err := catalog.NewClient(def, keys, httpClient)
		if err != nil {
			return // the fallback title stays; the run will report the key problem
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		title := requestTitle(ctx, client, ref.Model, system, firstMessage)
		if title != "" {
			app.Post(func() { c.session.Title = title; c.save() })
		}
	}()
}

func requestTitle(ctx context.Context, client llm.Client, model, system, message string) string {
	var reply strings.Builder
	req := llm.Request{
		Model: model, System: system, MaxTokens: 400,
		Messages: []llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart(truncateRunes(message, 2000))}}},
	}
	err := client.Stream(ctx, req, func(ev llm.Event) {
		if ev.Kind == llm.EventText {
			reply.WriteString(ev.Text)
		}
	})
	if err != nil {
		return ""
	}
	return truncateRunes(strings.Trim(strings.TrimSpace(reply.String()), `"'.`), maxTitleRunes)
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
