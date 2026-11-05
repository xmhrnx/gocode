package tools

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newFetchServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/page", func(w http.ResponseWriter, r *http.Request) {
		if r.UserAgent() != "GoCode/0.1" {
			http.Error(w, "bad agent", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("<html><head><script>var x = 1;</script></head><body><h1>Title</h1><p>Hello &amp; welcome</p></body></html>"))
	})
	mux.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	mux.HandleFunc("/image", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("\x89PNG"))
	})
	mux.Handle("/moved", http.RedirectHandler("/data", http.StatusFound))
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

func TestWebFetch(t *testing.T) {
	server := newFetchServer(t)
	env := Env{}
	tests := []struct {
		name    string
		args    webFetchArgs
		want    string
		wantErr bool
	}{
		{"html", webFetchArgs{URL: server.URL + "/page"}, "# Title\n\nHello & welcome", false},
		{"json via redirect", webFetchArgs{URL: server.URL + "/moved"}, "/data\nContent-Type: application/json\n\n{\"ok\":true}", false},
		{"max bytes", webFetchArgs{URL: server.URL + "/data", MaxBytes: 3}, "{\"o\n[truncated: 8 more bytes]", false},
		{"image refused", webFetchArgs{URL: server.URL + "/image"}, "unsupported content type", true},
		{"not found", webFetchArgs{URL: server.URL + "/missing"}, "HTTP 404", true},
		{"bad scheme", webFetchArgs{URL: "ftp://example.com"}, "http or https", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := runTool(t, WebFetch{}, env, tt.args)
			if res.IsError != tt.wantErr || !strings.Contains(res.Output, tt.want) {
				t.Errorf("got %q (error %v), want it to contain %q", res.Output, res.IsError, tt.want)
			}
		})
	}
}
