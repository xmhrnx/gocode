package settings

// WebSearch sets up how the agent searches the web.
type WebSearch struct {
	// Service is the search service of the web_search tool ("brave",
	// "tavily", "exa", "searxng"), or "" for none. Its key is in the
	// keychain.
	Service string `json:"service"`
	// SearXNGURL is the address of the SearXNG instance.
	SearXNGURL string `json:"searxng_url"`
	// Results is how many results a search returns.
	Results int `json:"results"`
	// ProviderSearch lets providers with a search of their own (Anthropic,
	// Gemini 3, OpenRouter) use it instead of the web_search tool.
	ProviderSearch bool `json:"provider_search"`
}
