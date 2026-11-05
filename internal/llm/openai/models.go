package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"gocode/internal/llm"
	"gocode/internal/llm/httpapi"
)

// modelEntry accepts the model shapes of OpenAI ({id}), OpenRouter
// ({id, name, context_length}), vLLM (max_model_len), Groq
// (context_window) and Ollama's native list ({name, model}).
type modelEntry struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Model         string `json:"model"`
	ContextLength int    `json:"context_length"`
	ContextWindow int    `json:"context_window"`
	MaxModelLen   int    `json:"max_model_len"`
}

type modelList struct {
	Data   []modelEntry `json:"data"`
	Models []modelEntry `json:"models"`
}

// ListModels implements llm.Client.
func (c *Client) ListModels(ctx context.Context) ([]llm.ModelInfo, error) {
	var raw json.RawMessage
	if err := httpapi.GetJSON(ctx, c.http, c.opts.BaseURL+"/models", c.header(), &raw); err != nil {
		return nil, err
	}
	entries, err := decodeModelEntries(raw)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var models []llm.ModelInfo
	for _, entry := range entries {
		info := entry.toLLM()
		if info.ID == "" || seen[info.ID] {
			continue
		}
		seen[info.ID] = true
		models = append(models, info)
	}
	slices.SortFunc(models, func(a, b llm.ModelInfo) int { return strings.Compare(a.ID, b.ID) })
	return models, nil
}

// decodeModelEntries accepts {data:[...]}, {models:[...]} or a bare array.
func decodeModelEntries(raw json.RawMessage) ([]modelEntry, error) {
	var entries []modelEntry
	if json.Unmarshal(raw, &entries) == nil {
		return entries, nil
	}
	var list modelList
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, fmt.Errorf("decode model list: %w", err)
	}
	return append(list.Data, list.Models...), nil
}

func (e modelEntry) toLLM() llm.ModelInfo {
	id := firstNonEmpty(e.ID, e.Model, e.Name)
	info := llm.ModelInfo{ID: id, ContextWindow: max(e.ContextLength, e.ContextWindow, e.MaxModelLen)}
	if e.Name != id {
		info.Name = e.Name
	}
	return info
}

// firstNonEmpty returns the first non-empty string.
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
