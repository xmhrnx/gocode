package providers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestCatalogEntriesAreWellFormed(t *testing.T) {
	seen := map[string]bool{}
	categories := map[string]bool{CategoryCloud: true, CategoryRouter: true, CategoryLocal: true}
	for _, def := range Catalog() {
		if def.ID == "" || def.Name == "" || seen[def.ID] {
			t.Errorf("missing or duplicate id/name: %+v", def)
		}
		seen[def.ID] = true
		if !categories[def.Category] {
			t.Errorf("%s: unknown category %q", def.ID, def.Category)
		}
		parsed, err := url.Parse(def.BaseURL)
		if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
			t.Errorf("%s: bad base URL %q (%v)", def.ID, def.BaseURL, err)
		}
		if def.Category == CategoryLocal != def.KeyOptional {
			t.Errorf("%s: only local servers should have optional keys", def.ID)
		}
		if _, err := NewClient(def, "key", nil); err != nil {
			t.Errorf("%s: NewClient: %v", def.ID, err)
		}
	}
	if len(seen) < 30 {
		t.Errorf("catalog has only %d entries", len(seen))
	}
}

func TestLookup(t *testing.T) {
	def, ok := Lookup("deepseek")
	if !ok || def.EchoReasoningField != "reasoning_content" || def.Style != StyleOpenAI {
		t.Errorf("Lookup(deepseek) = %+v, %v", def, ok)
	}
	if _, ok := Lookup("nope"); ok {
		t.Error("Lookup found an unknown provider")
	}
}

func TestNewClientPerStyle(t *testing.T) {
	var gotPath, gotHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotHeader = r.Header.Get("x-api-key") + r.Header.Get("x-goog-api-key") + r.Header.Get("Authorization") + r.Header.Get("X-Extra")
		_, _ = w.Write([]byte(`{"data":[],"models":[]}`))
	}))
	defer server.Close()
	cases := []struct {
		style      Style
		path, auth string
	}{
		{StyleOpenAI, "/models", "Bearer kyes"},
		{StyleAnthropic, "/models", "kyes"},
		{StyleGemini, "/models", "kyes"},
	}
	for _, c := range cases {
		def := Definition{ID: string(c.style), Style: c.style, BaseURL: server.URL, Headers: map[string]string{"X-Extra": "yes"}}
		client, err := NewClient(def, "k", server.Client())
		if err != nil {
			t.Fatalf("%s: %v", c.style, err)
		}
		if _, err := client.ListModels(context.Background()); err != nil {
			t.Errorf("%s: ListModels: %v", c.style, err)
		}
		if gotPath != c.path || gotHeader != c.auth {
			t.Errorf("%s: path %q auth %q", c.style, gotPath, gotHeader)
		}
	}
	if _, err := NewClient(Definition{ID: "x", Style: "soap"}, "", nil); err == nil {
		t.Error("unknown style accepted")
	}
	if _, err := NewClient(Definition{ID: "x", Style: StyleOpenAI}, "", nil); err == nil {
		t.Error("openai style without base URL accepted")
	}
}

func TestDirectoryProvidersAreUsable(t *testing.T) {
	seen := map[string]bool{}
	for _, def := range Catalog() {
		if seen[def.ID] {
			t.Errorf("duplicate provider id %q", def.ID)
		}
		seen[def.ID] = true
		if def.Name == "" || def.BaseURL == "" || (def.Style != StyleOpenAI && def.Style != StyleAnthropic && def.Style != StyleGemini) {
			t.Errorf("provider %q is incomplete: %+v", def.ID, def)
		}
		if def.Category != CategoryCloud && def.Category != CategoryRouter && def.Category != CategoryLocal {
			t.Errorf("provider %q has category %q", def.ID, def.Category)
		}
	}
	if len(seen) < 200 {
		t.Fatalf("only %d providers", len(seen))
	}
}
