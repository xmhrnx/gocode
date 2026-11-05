package anthropic

import (
	"context"
	"slices"
	"strings"

	"gocode/internal/llm"
	"gocode/internal/llm/httpapi"
)

type modelList struct {
	Data []struct {
		ID             string `json:"id"`
		DisplayName    string `json:"display_name"`
		MaxInputTokens int    `json:"max_input_tokens"`
	} `json:"data"`
}

// ListModels implements llm.Client.
func (c *Client) ListModels(ctx context.Context) ([]llm.ModelInfo, error) {
	var list modelList
	if err := httpapi.GetJSON(ctx, c.http, c.opts.BaseURL+"/models?limit=1000", c.header(), &list); err != nil {
		return nil, err
	}
	models := make([]llm.ModelInfo, 0, len(list.Data))
	for _, entry := range list.Data {
		models = append(models, llm.ModelInfo{ID: entry.ID, Name: entry.DisplayName, ContextWindow: entry.MaxInputTokens})
	}
	slices.SortFunc(models, func(a, b llm.ModelInfo) int { return strings.Compare(a.ID, b.ID) })
	return models, nil
}
