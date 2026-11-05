package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gocode/internal/llm"
)

func TestErrorMessageShapes(t *testing.T) {
	cases := map[string]string{
		`{"error":{"message":"bad key","type":"auth"}}`: "bad key",
		`{"error":"plain"}`:                             "plain",
		`{"message":"top level"}`:                       "top level",
		"  <html>oops</html> ":                          "<html>oops</html>",
		"":                                              "fallback",
	}
	for body, want := range cases {
		if got := ErrorMessage([]byte(body), "fallback"); got != want {
			t.Errorf("ErrorMessage(%q) = %q, want %q", body, got, want)
		}
	}
	long := ErrorMessage([]byte(strings.Repeat("é", 400)), "")
	if len(long) > maxMessageLen+len("…") || !strings.HasSuffix(long, "…") {
		t.Errorf("long message not truncated: %d bytes", len(long))
	}
}

func TestSendReturnsAPIErrorOnFailureStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Test") != "yes" || r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("headers not sent: %v", r.Header)
		}
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":{"message":"slow down"}}`))
	}))
	defer server.Close()
	header := http.Header{"X-Test": {"yes"}}
	_, err := Send(context.Background(), Client(nil), http.MethodPost, server.URL, header, map[string]int{"a": 1})
	var apiErr *llm.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 429 || apiErr.Message != "slow down" {
		t.Fatalf("err = %v", err)
	}
}

func TestGetJSONDecodes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"value":3}`))
	}))
	defer server.Close()
	var out struct{ Value int }
	if err := GetJSON(context.Background(), server.Client(), server.URL, http.Header{}, &out); err != nil || out.Value != 3 {
		t.Fatalf("GetJSON = %v, %+v", err, out)
	}
}
