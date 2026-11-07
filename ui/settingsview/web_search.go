package settingsview

import (
	"strings"

	"gocode/internal/websearch"
	"gocode/ui/core"
)

var searchServices = []option[string]{
	{"", "Off"},
	{websearch.ServiceBrave, "Brave Search"},
	{websearch.ServiceTavily, "Tavily"},
	{websearch.ServiceExa, "Exa"},
	{websearch.ServiceSearXNG, "SearXNG (your own)"},
}

func serviceName(service string) string {
	for _, o := range searchServices {
		if o.value == service {
			return strings.TrimSuffix(o.label, " (your own)")
		}
	}
	return service
}

func webSearchEntries(s *core.Services) []entry {
	w := &s.Settings.WebSearch
	return []entry{
		{title: "Search service", description: "Gives the agent a web_search tool. Brave (api-dashboard.search.brave.com), Tavily (tavily.com) and Exa (exa.ai) need an API key and have free tiers; SearXNG needs only its address.",
			keywords: "web search internet online google brave tavily exa searxng lookup", control: choice(s, &w.Service, searchServices)},
		{title: "API key", description: "The key for the service chosen above, kept in the system keychain. Test runs one search.",
			keywords: "web search api key token brave tavily exa", control: newSearchKeyControl(s)},
		{title: "SearXNG address", description: "Your instance, e.g. http://localhost:8080. It must allow the JSON format (search.formats in settings.yml).",
			keywords: "searxng url address instance self-hosted", control: text(s, &w.SearXNGURL, "http://localhost:8080", 260)},
		{title: "Results per search", description: "How many results a search returns to the model.",
			keywords: "web search results count number", control: number(s, &w.Results, 1, 20)},
		{title: "Use the provider's own search", description: "Anthropic, Gemini 3 and OpenRouter models search the web themselves instead of using the service above; other models keep the service. The provider bills each search (Anthropic: $10 per 1,000).",
			keywords: "native provider web search anthropic claude gemini google grounding openrouter", control: switchControl(s, &w.ProviderSearch)},
	}
}
