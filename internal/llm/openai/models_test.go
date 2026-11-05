package openai

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gocode/internal/llm"
)

func serveJSON(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" && r.URL.Path != "/chat/completions" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return server
}

func TestListModelsShapes(t *testing.T) {
	cases := map[string]string{
		"openai":     `{"object":"list","data":[{"id":"b"},{"id":"a","context_length":8192}]}`,
		"ollama":     `{"models":[{"name":"b","model":"b"},{"name":"a","model":"a"}]}`,
		"bare array": `[{"id":"b"},{"id":"a"},{"id":"a"}]`,
	}
	for name, body := range cases {
		models, err := New(Options{BaseURL: serveJSON(t, 200, body).URL}).ListModels(context.Background())
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if len(models) != 2 || models[0].ID != "a" || models[1].ID != "b" || models[0].Name != "" {
			t.Errorf("%s: models = %+v", name, models)
		}
	}
}

func TestHTTPErrorsBecomeAPIErrors(t *testing.T) {
	server := serveJSON(t, 401, `{"error":{"message":"Incorrect API key provided","type":"invalid_request_error"}}`)
	client := New(Options{BaseURL: server.URL})
	_, err := streamAll(t, client, llm.Request{Model: "m"})
	var apiErr *llm.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 401 || apiErr.Message != "Incorrect API key provided" {
		t.Errorf("stream error = %v", err)
	}
	if _, err := client.ListModels(context.Background()); !errors.As(err, &apiErr) {
		t.Errorf("models error = %v", err)
	}
}
