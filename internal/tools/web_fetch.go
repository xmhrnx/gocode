package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"gocode/internal/llm"
)

// WebFetch downloads a web page as text.
type WebFetch struct {
	// Client sends the requests; nil uses a client with a 30s timeout.
	Client *http.Client
}

type webFetchArgs struct {
	URL      string `json:"url"`
	MaxBytes int    `json:"max_bytes"`
}

const (
	fetchTimeout   = 30 * time.Second
	fetchUserAgent = "GoCode/0.1"
	// maxFetchBodyBytes bounds how much of a response is downloaded before
	// conversion; pages larger than this are cut, not refused.
	maxFetchBodyBytes = 5 << 20
)

// Spec describes web_fetch to the model.
func (WebFetch) Spec() llm.ToolSpec {
	return llm.ToolSpec{
		Name: "web_fetch",
		Description: "Fetch a URL with an HTTP GET (following redirects, 30 second timeout) and return its content as text. " +
			"HTML is converted to readable plain text (scripts and styles removed); JSON, XML and plain text are returned as-is. " +
			"Other content types such as images and PDFs are refused. Use it to read documentation or API responses.",
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "url": {"type": "string", "description": "Absolute http or https URL to fetch."},
    "max_bytes": {"type": "integer", "minimum": 1, "description": "Maximum bytes of text to return; defaults to and is capped at the tool output limit (30000 by default)."}
  },
  "required": ["url"],
  "additionalProperties": false
}`),
	}
}

// Access reports that web_fetch reaches the network.
func (WebFetch) Access() Access { return AccessNetwork }

// Summary names the URL.
func (WebFetch) Summary(raw json.RawMessage) string {
	var args webFetchArgs
	summaryArgs(raw, &args)
	return "Fetch " + shorten(args.URL, 100)
}

// Run fetches the page.
func (t WebFetch) Run(ctx context.Context, env Env, raw json.RawMessage) Result {
	var args webFetchArgs
	if err := decodeArgs(raw, &args); err != nil {
		return Errorf("%v", err)
	}
	target, err := url.Parse(args.URL)
	if err != nil || (target.Scheme != "http" && target.Scheme != "https") || target.Host == "" {
		return Errorf("url must be an absolute http or https URL, got %q", args.URL)
	}
	resp, err := t.get(ctx, target.String())
	if err != nil {
		return Errorf("fetching %s failed: %v", args.URL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return Errorf("fetching %s failed: HTTP %s", args.URL, resp.Status)
	}
	text, contentType, err := responseText(resp, io.LimitReader(resp.Body, maxFetchBodyBytes))
	if err != nil {
		return Errorf("%s: %v", args.URL, err)
	}
	limit := outputLimit(env)
	if args.MaxBytes > 0 {
		limit = min(limit, args.MaxBytes)
	}
	header := fmt.Sprintf("URL: %s\nContent-Type: %s\n\n", resp.Request.URL, contentType)
	return Result{Output: header + truncateHead(text, limit)}
}

// get sends the GET request with GoCode's headers.
func (t WebFetch) get(ctx context.Context, target string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", fetchUserAgent)
	req.Header.Set("Accept", "text/html, application/xhtml+xml, application/json, text/plain;q=0.9, */*;q=0.5")
	client := t.Client
	if client == nil {
		client = &http.Client{Timeout: fetchTimeout}
	}
	return client.Do(req)
}
