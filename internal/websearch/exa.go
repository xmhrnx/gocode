package websearch

import (
	"context"
	"net/http"
)

const exaURL = "https://api.exa.ai/search"

// exaSnippetChars is how much of each page's text Exa returns.
const exaSnippetChars = 400

// exa searches with the Exa API.
type exa struct {
	key, baseURL string
	client       *http.Client
}

type exaRequest struct {
	Query      string      `json:"query"`
	NumResults int         `json:"numResults"`
	Contents   exaContents `json:"contents"`
}

type exaContents struct {
	Text exaText `json:"text"`
}

type exaText struct {
	MaxCharacters int `json:"maxCharacters"`
}

type exaResponse struct {
	Results []struct {
		Title string `json:"title"`
		URL   string `json:"url"`
		Text  string `json:"text"`
	} `json:"results"`
}

func (e exa) Search(ctx context.Context, query string, count int) ([]Result, error) {
	body := exaRequest{Query: query, NumResults: count, Contents: exaContents{Text: exaText{MaxCharacters: exaSnippetChars}}}
	var resp exaResponse
	if err := doJSON(ctx, e.client, http.MethodPost, e.baseURL, map[string]string{"x-api-key": e.key}, body, &resp); err != nil {
		return nil, err
	}
	results := make([]Result, 0, len(resp.Results))
	for _, r := range resp.Results {
		results = append(results, Result{Title: plain(r.Title), URL: r.URL, Snippet: plain(r.Text)})
	}
	return results, nil
}
