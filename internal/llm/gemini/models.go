package gemini

import (
	"context"
	"net/url"
	"slices"
	"strings"

	"gocode/internal/llm"
	"gocode/internal/llm/httpapi"
)

type modelPage struct {
	Models []struct {
		Name                       string   `json:"name"`
		DisplayName                string   `json:"displayName"`
		InputTokenLimit            int      `json:"inputTokenLimit"`
		SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
	} `json:"models"`
	NextPageToken string `json:"nextPageToken"`
}

// ListModels implements llm.Client. It returns only models that can
// generate content, skipping embedding and other special-purpose models.
func (c *Client) ListModels(ctx context.Context) ([]llm.ModelInfo, error) {
	var models []llm.ModelInfo
	pageToken := ""
	for {
		query := url.Values{"pageSize": {"1000"}}
		if pageToken != "" {
			query.Set("pageToken", pageToken)
		}
		var page modelPage
		if err := httpapi.GetJSON(ctx, c.http, c.opts.BaseURL+"/models?"+query.Encode(), c.header(), &page); err != nil {
			return nil, err
		}
		for _, model := range page.Models {
			if slices.Contains(model.SupportedGenerationMethods, "generateContent") {
				models = append(models, llm.ModelInfo{
					ID:            strings.TrimPrefix(model.Name, "models/"),
					Name:          model.DisplayName,
					ContextWindow: model.InputTokenLimit,
				})
			}
		}
		if page.NextPageToken == "" {
			break
		}
		pageToken = page.NextPageToken
	}
	slices.SortFunc(models, func(a, b llm.ModelInfo) int { return strings.Compare(a.ID, b.ID) })
	return models, nil
}
