package websearch

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

const braveURL = "https://api.search.brave.com/res/v1/web/search"

// brave searches with the Brave Search API.
type brave struct {
	key, baseURL string
	client       *http.Client
}

type braveResponse struct {
	Web struct {
		Results []struct {
			Title       string `json:"title"`
			URL         string `json:"url"`
			Description string `json:"description"`
		} `json:"results"`
	} `json:"web"`
}

func (b brave) Search(ctx context.Context, query string, count int) ([]Result, error) {
	params := url.Values{"q": {query}, "count": {strconv.Itoa(count)}}
	var resp braveResponse
	headers := map[string]string{"X-Subscription-Token": b.key}
	if err := doJSON(ctx, b.client, http.MethodGet, b.baseURL+"?"+params.Encode(), headers, nil, &resp); err != nil {
		return nil, err
	}
	results := make([]Result, 0, len(resp.Web.Results))
	for _, r := range resp.Web.Results {
		results = append(results, Result{Title: plain(r.Title), URL: r.URL, Snippet: plain(r.Description)})
	}
	return results, nil
}
