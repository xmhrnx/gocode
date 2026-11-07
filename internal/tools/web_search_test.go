package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"gocode/internal/websearch"
)

type fakeSearcher struct {
	results []websearch.Result
	err     error
	count   int
}

func (f *fakeSearcher) Search(_ context.Context, _ string, count int) ([]websearch.Result, error) {
	f.count = count
	return f.results, f.err
}

func TestWebSearchListsResults(t *testing.T) {
	searcher := &fakeSearcher{results: []websearch.Result{{Title: "Go", URL: "https://go.dev", Snippet: "The language"}}}
	tool := WebSearch{Searcher: searcher, Results: 5}
	result := tool.Run(context.Background(), Env{}, json.RawMessage(`{"query":"go","count":9}`))
	if result.IsError || !strings.Contains(result.Output, "1. Go\n   https://go.dev\n   The language") || searcher.count != 5 {
		t.Fatalf("output %q, count %d", result.Output, searcher.count)
	}
	searcher.err = errors.New("401 bad key")
	if result := tool.Run(context.Background(), Env{}, json.RawMessage(`{"query":"go"}`)); !result.IsError || !strings.Contains(result.Output, "bad key") {
		t.Fatalf("error result %+v", result)
	}
	if result := tool.Run(context.Background(), Env{}, json.RawMessage(`{"query":" "}`)); !result.IsError {
		t.Fatal("an empty query must fail")
	}
}
