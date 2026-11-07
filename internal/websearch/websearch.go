// Package websearch looks things up on the web through a search service
// the user has a key for (or a SearXNG instance), for the agent's
// web_search tool.
package websearch

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// Result is one search hit.
type Result struct {
	Title   string
	URL     string
	Snippet string
}

// Searcher runs a web search.
type Searcher interface {
	Search(ctx context.Context, query string, count int) ([]Result, error)
}

// Config chooses and sets up a search service.
type Config struct {
	// Service is one of the Service* names.
	Service string
	Key     string
	// URL is the SearXNG instance's address.
	URL    string
	Client *http.Client
}

// The search services.
const (
	ServiceBrave   = "brave"
	ServiceTavily  = "tavily"
	ServiceExa     = "exa"
	ServiceSearXNG = "searxng"
)

// NeedsKey reports whether service needs an API key.
func NeedsKey(service string) bool { return service != ServiceSearXNG && service != "" }

// KeyName is the keychain entry holding a service's key.
func KeyName(service string) string { return "websearch." + service + ".key" }

// New returns a searcher for cfg.
func New(cfg Config) (Searcher, error) {
	client := cfg.Client
	if client == nil {
		client = http.DefaultClient
	}
	if NeedsKey(cfg.Service) && cfg.Key == "" {
		return nil, fmt.Errorf("no API key saved for %s", cfg.Service)
	}
	switch cfg.Service {
	case ServiceBrave:
		return brave{key: cfg.Key, client: client, baseURL: braveURL}, nil
	case ServiceTavily:
		return tavily{key: cfg.Key, client: client, baseURL: tavilyURL}, nil
	case ServiceExa:
		return exa{key: cfg.Key, client: client, baseURL: exaURL}, nil
	case ServiceSearXNG:
		if strings.TrimSpace(cfg.URL) == "" {
			return nil, fmt.Errorf("no SearXNG address set")
		}
		return searxng{client: client, baseURL: strings.TrimRight(strings.TrimSpace(cfg.URL), "/")}, nil
	}
	return nil, fmt.Errorf("unknown search service %q", cfg.Service)
}
