package websearch

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fake answers with body and records what it was asked.
func fake(t *testing.T, body string, seen *http.Request, sent *map[string]any) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*seen = *r.Clone(context.Background())
		if data, _ := io.ReadAll(r.Body); len(data) > 0 {
			_ = json.Unmarshal(data, sent)
		}
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(server.Close)
	return server
}

func TestServices(t *testing.T) {
	cases := []struct {
		name, body string
		make       func(url string) Searcher
		check      func(t *testing.T, r *http.Request, sent map[string]any)
	}{
		{"brave", `{"web":{"results":[{"title":"Go <strong>1.26</strong>","url":"https://go.dev","description":"The &amp; release"}]}}`,
			func(u string) Searcher { return brave{key: "k", client: http.DefaultClient, baseURL: u} },
			func(t *testing.T, r *http.Request, _ map[string]any) {
				if r.Header.Get("X-Subscription-Token") != "k" || r.URL.Query().Get("q") != "go" || r.URL.Query().Get("count") != "3" {
					t.Errorf("brave request %v %v", r.Header, r.URL)
				}
			}},
		{"tavily", `{"results":[{"title":"Go 1.26","url":"https://go.dev","content":"The & release"}]}`,
			func(u string) Searcher { return tavily{key: "k", client: http.DefaultClient, baseURL: u} },
			func(t *testing.T, r *http.Request, sent map[string]any) {
				if r.Header.Get("Authorization") != "Bearer k" || sent["query"] != "go" || sent["max_results"] != 3.0 {
					t.Errorf("tavily request %v %v", r.Header, sent)
				}
			}},
		{"exa", `{"results":[{"title":"Go 1.26","url":"https://go.dev","text":"The & release"}]}`,
			func(u string) Searcher { return exa{key: "k", client: http.DefaultClient, baseURL: u} },
			func(t *testing.T, r *http.Request, sent map[string]any) {
				if r.Header.Get("x-api-key") != "k" || sent["query"] != "go" || sent["numResults"] != 3.0 {
					t.Errorf("exa request %v %v", r.Header, sent)
				}
			}},
		{"searxng", `{"results":[{"title":"Go 1.26","url":"https://go.dev","content":"The & release"},{"title":"x"},{"title":"y"},{"title":"z"}]}`,
			func(u string) Searcher { return searxng{client: http.DefaultClient, baseURL: u} },
			func(t *testing.T, r *http.Request, _ map[string]any) {
				if r.URL.Path != "/search" || r.URL.Query().Get("format") != "json" {
					t.Errorf("searxng request %v", r.URL)
				}
			}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var seen http.Request
			sent := map[string]any{}
			server := fake(t, c.body, &seen, &sent)
			results, err := c.make(server.URL).Search(context.Background(), "go", 3)
			if err != nil {
				t.Fatal(err)
			}
			c.check(t, &seen, sent)
			if len(results) == 0 || len(results) > 3 || results[0].Title != "Go 1.26" || results[0].URL != "https://go.dev" || results[0].Snippet != "The & release" {
				t.Fatalf("results %+v", results)
			}
		})
	}
}

func TestErrorsAndConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":"bad key"}`, http.StatusUnauthorized)
	}))
	defer server.Close()
	_, err := brave{key: "k", client: http.DefaultClient, baseURL: server.URL}.Search(context.Background(), "go", 3)
	if err == nil || !strings.Contains(err.Error(), "401") || !strings.Contains(err.Error(), "bad key") {
		t.Fatalf("error %v", err)
	}
	if _, err := New(Config{Service: ServiceBrave}); err == nil {
		t.Fatal("a missing key must be reported")
	}
	if _, err := New(Config{Service: ServiceSearXNG, URL: "http://localhost:8080/"}); err != nil {
		t.Fatal(err)
	}
}
