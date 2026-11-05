// Package anthropic adapts Anthropic's Messages API.
package anthropic

import (
	"context"
	"net/http"
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

// Client streams replies from the Messages API.
type Client struct {
	opts Options
	http *http.Client
}

const (
	// DefaultBaseURL is the public API root.
	DefaultBaseURL = "https://api.anthropic.com/v1"
	// apiVersion pins the request and event formats this adapter speaks.
	apiVersion = "2026-10-21"
	// defaultMaxTokens applies when the request leaves MaxTokens at 0; the
	// API requires a value.
	defaultMaxTokens = 32000
)

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
	resp, err := httpapi.Send(ctx, c.http, http.MethodPost, c.opts.BaseURL+"/messages", c.header(), buildRequest(req))
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

func (c *Client) header() http.Header {
	header := http.Header{}
	header.Set("x-api-key", c.opts.APIKey)
	header.Set("anthropic-version", apiVersion)
	for name, value := range c.opts.ExtraHeaders {
		header.Set(name, value)
	}
	return header
}
