// Package openai adapts the Chat Completions API, which OpenAI and almost
// every OpenAI-compatible provider and local server speak.
package openai

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
	// BaseURL is the API root, e.g. "https://api.openai.com/v1".
	BaseURL string
	APIKey  string
	// AuthHeader names the header that carries the key. "" or
	// "Authorization" sends "Bearer <key>"; any other name (e.g. Azure's
	// "api-key") sends the raw key.
	AuthHeader   string
	ExtraHeaders map[string]string
	// HTTPClient defaults to http.DefaultClient.
	HTTPClient *http.Client
	// MaxTokensField is MaxCompletionTokens (default) or MaxTokens; many
	// compatible servers only understand the older "max_tokens".
	MaxTokensField string
	// EchoReasoningField, when set (e.g. "reasoning_content"), sends an
	// assistant message's reasoning back in that field. DeepSeek, Kimi and
	// Z.ai require it during tool loops.
	EchoReasoningField string
	// SendReasoningEffort sends Request.ReasoningEffort as
	// "reasoning_effort"; servers that don't know it may reject it.
	SendReasoningEffort bool
	// WebSearchTool is the type of the provider's own web search tool,
	// sent when a request asks for web search; "" when there is none.
	WebSearchTool string
	// FileInput sends attached PDFs as files ("file" content parts),
	// which OpenAI and OpenRouter read; without it they go as their text.
	FileInput bool
	// AudioInput sends attached WAV and MP3 files as input_audio, which
	// OpenAI's and OpenRouter's audio models take.
	AudioInput bool
}

// Client streams chat completions from one provider.
type Client struct {
	opts Options
	http *http.Client
}

// Names accepted by Options.MaxTokensField.
const (
	MaxCompletionTokens = "max_completion_tokens"
	MaxTokens           = "max_tokens"
)

// New returns a Client for the given options.
func New(opts Options) *Client {
	if opts.MaxTokensField == "" {
		opts.MaxTokensField = MaxCompletionTokens
	}
	opts.BaseURL = strings.TrimRight(opts.BaseURL, "/")
	return &Client{opts: opts, http: httpapi.Client(opts.HTTPClient)}
}

// Stream implements llm.Client.
func (c *Client) Stream(ctx context.Context, req llm.Request, emit func(llm.Event)) error {
	url := c.opts.BaseURL + "/chat/completions"
	resp, err := httpapi.Send(ctx, c.http, http.MethodPost, url, c.header(), c.buildRequest(req))
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
	if c.opts.APIKey != "" {
		if c.opts.AuthHeader == "" || strings.EqualFold(c.opts.AuthHeader, "Authorization") {
			header.Set("Authorization", "Bearer "+c.opts.APIKey)
		} else {
			header.Set(c.opts.AuthHeader, c.opts.APIKey)
		}
	}
	for name, value := range c.opts.ExtraHeaders {
		header.Set(name, value)
	}
	return header
}
