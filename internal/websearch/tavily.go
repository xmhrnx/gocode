package websearch

import (
	"context"
	"net/http"
)

const tavilyURL = "https://api.tavily.com/search"

// tavily searches with the Tavily API.
type tavily struct {
	key, baseURL string
	client       *http.Client
}

type tavilyRequest struct {
	Query      string `json:"query"`
	MaxResults int    `json:"max_results"`
}

type tavilyResponse struct {
	Results []struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Content string `json:"content"`
	} `json:"results"`
}

func (t tavily) Search(ctx context.Context, query string, count int) ([]Result, error) {
	var resp tavilyResponse
	headers := map[string]string{"Authorization": "Bearer " + t.key}
	if err := doJSON(ctx, t.client, http.MethodPost, t.baseURL, headers, tavilyRequest{Query: query, MaxResults: count}, &resp); err != nil {
		return nil, err
	}
	results := make([]Result, 0, len(resp.Results))
	for _, r := range resp.Results {
		results = append(results, Result{Title: plain(r.Title), URL: r.URL, Snippet: plain(r.Content)})
	}
	return results, nil
}
