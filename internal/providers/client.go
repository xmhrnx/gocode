package providers

import (
	"errors"
	"fmt"
	"net/http"

	"gocode/internal/llm"
	"gocode/internal/llm/anthropic"
	"gocode/internal/llm/gemini"
	"gocode/internal/llm/openai"
)

// NewClient returns a client for def's API, authenticated with apiKey. A
// nil httpClient uses http.DefaultClient.
func NewClient(def Definition, apiKey string, httpClient *http.Client) (llm.Client, error) {
	switch def.Style {
	case StyleAnthropic:
		return anthropic.New(anthropic.Options{
			BaseURL: def.BaseURL, APIKey: apiKey, ExtraHeaders: def.Headers, HTTPClient: httpClient,
		}), nil
	case StyleGemini:
		return gemini.New(gemini.Options{
			BaseURL: def.BaseURL, APIKey: apiKey, ExtraHeaders: def.Headers, HTTPClient: httpClient,
		}), nil
	case StyleOpenAI:
		if def.BaseURL == "" {
			return nil, errors.New("provider " + def.ID + " has no base URL")
		}
		return openai.New(openai.Options{
			BaseURL:             def.BaseURL,
			APIKey:              apiKey,
			AuthHeader:          def.AuthHeader,
			ExtraHeaders:        def.Headers,
			HTTPClient:          httpClient,
			MaxTokensField:      def.MaxTokensField,
			EchoReasoningField:  def.EchoReasoningField,
			SendReasoningEffort: def.SendReasoningEffort,
			WebSearchTool:       def.WebSearchTool,
		}), nil
	}
	return nil, fmt.Errorf("provider %s has unknown API style %q", def.ID, def.Style)
}
