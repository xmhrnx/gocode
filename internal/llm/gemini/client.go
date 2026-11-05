// Package gemini adapts Google's native Gemini API (generateContent).
package gemini

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"gocode/internal/llm"
	"gocode/internal/llm/httpapi"
	"gocode/internal/llm/sse"
)

// Options configures a Client.
type Options struct {
	// BaseURL defaults to DefaultBaseURL.
	BaseURL      string
	APIKey       string
	ExtraHeaders map[string]string
	// HTTPClient defaults to http.DefaultClient.
	HTTPClient *http.Client
}

// Client streams replies from the Gemini API.
type Client struct {
	opts Options
	http *http.Client
}

// DefaultBaseURL is the public API root.
const DefaultBaseURL = "https://generativelanguage.googleapis.com/v1beta"

// New returns a Client for the given options.
func New(opts Options) *Client {
	if opts.BaseURL == "" {
		opts.BaseURL = DefaultBaseURL
	}
	opts.BaseURL = strings.TrimRight(opts.BaseURL, "/")
	return &Client{opts: opts, http: httpapi.Client(opts.HTTPClient)}
}

// Stream implements llm.Client.
func (c *Client) Stream(ctx context.Context, req llm.Request, emit func(llm.Event)) error {
	resp, err := httpapi.Send(ctx, c.http, http.MethodPost, c.streamURL(req.Model), c.header(), buildRequest(req))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	stream := newStreamState(emit)
	if err := sse.ForEach(resp.Body, stream.handle); err != nil {
		return err
	}
	return stream.finish()
}

func (c *Client) streamURL(model string) string {
	model = strings.TrimPrefix(model, "models/")
	return c.opts.BaseURL + "/models/" + url.PathEscape(model) + ":streamGenerateContent?alt=sse"
}

func (c *Client) header() http.Header {
	header := http.Header{}
	header.Set("x-goog-api-key", c.opts.APIKey)
	for name, value := range c.opts.ExtraHeaders {
		header.Set(name, value)
	}
	return header
}
