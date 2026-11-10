// Package providers lists the model providers GoCode knows how to reach
// and builds an llm.Client for any of them.
package providers

// Style is the wire protocol a provider speaks.
type Style string

// Definition describes one provider: where it lives and the quirks of its
// API. Fields beyond ID, Name, Style and BaseURL only matter to the
// OpenAI-compatible adapter.
type Definition struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Style    Style  `json:"style"`
	BaseURL  string `json:"base_url"`
	Category string `json:"category"`
	// KeyURL is where to create an API key, or "" when unknown.
	KeyURL string `json:"key_url,omitempty"`
	// KeyOptional is true for servers that accept requests without a key.
	KeyOptional bool `json:"key_optional,omitempty"`
	// AuthHeader is "" for "Authorization: Bearer <key>", or a header name
	// (e.g. "api-key") that carries the raw key.
	AuthHeader string            `json:"auth_header,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	// MaxTokensField is "max_tokens" for servers that don't understand
	// "max_completion_tokens" (the default).
	MaxTokensField string `json:"max_tokens_field,omitempty"`
	// EchoReasoningField names the assistant-message field in which the
	// provider expects earlier reasoning back, e.g. "reasoning_content".
	EchoReasoningField  string `json:"echo_reasoning_field,omitempty"`
	SendReasoningEffort bool   `json:"send_reasoning_effort,omitempty"`
	// WebSearchTool names the provider's own web search tool, if it has
	// one for every model (OpenRouter's "openrouter:web_search").
	WebSearchTool string `json:"web_search_tool,omitempty"`
	// AudioInput marks a provider whose audio models take WAV and MP3.
	AudioInput bool `json:"audio_input,omitempty"`
	// FileInput marks a provider that reads PDFs sent as files.
	FileInput       bool     `json:"file_input,omitempty"`
	SuggestedModels []string `json:"suggested_models,omitempty"`
	Notes           string   `json:"notes,omitempty"`
}

// Supported styles.
const (
	StyleOpenAI    Style = "openai"
	StyleAnthropic Style = "anthropic"
	StyleGemini    Style = "gemini"
)

// Categories group providers for display.
const (
	CategoryCloud  = "cloud"  // model labs' own APIs
	CategoryRouter = "router" // routers and inference hosts
	CategoryLocal  = "local"  // servers running on this machine
)
