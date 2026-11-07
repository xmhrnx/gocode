package websearch

import (
	"context"
	"net/http"
	"net/url"
)

// searxng searches a SearXNG instance, which must allow the JSON format
// (search.formats in its settings.yml).
type searxng struct {
	baseURL string
	client  *http.Client
}

type searxngResponse struct {
	Results []struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Content string `json:"content"`
	} `json:"results"`
}

func (s searxng) Search(ctx context.Context, query string, count int) ([]Result, error) {
	params := url.Values{"q": {query}, "format": {"json"}}
	var resp searxngResponse
	if err := doJSON(ctx, s.client, http.MethodGet, s.baseURL+"/search?"+params.Encode(), nil, nil, &resp); err != nil {
		return nil, err
	}
	results := make([]Result, 0, min(len(resp.Results), count))
	for _, r := range resp.Results[:min(len(resp.Results), count)] {
		results = append(results, Result{Title: plain(r.Title), URL: r.URL, Snippet: plain(r.Content)})
	}
	return results, nil
}
