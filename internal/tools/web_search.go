package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gocode/internal/llm"
	"gocode/internal/websearch"
)

// WebSearch looks things up with the search service set in Settings.
type WebSearch struct {
	Searcher websearch.Searcher
	// Results is how many results a search returns unless the model asks
	// for fewer.
	Results int
}

type webSearchArgs struct {
	Query string `json:"query"`
	Count int    `json:"count"`
}

const searchTimeout = 30 * time.Second

// Spec describes web_search to the model.
func (t WebSearch) Spec() llm.ToolSpec {
	return llm.ToolSpec{
		Name: llm.WebSearchTool,
		Description: "Search the web. Returns titles, URLs and short snippets; read a promising page in full with web_fetch. " +
			"Use it for current information, documentation, error messages and anything that may have changed since your training.",
		Parameters: json.RawMessage(fmt.Sprintf(`{
  "type": "object",
  "properties": {
    "query": {"type": "string", "description": "What to search for, as you would type it into a search engine."},
    "count": {"type": "integer", "minimum": 1, "maximum": %d, "description": "How many results to return; defaults to %d."}
  },
  "required": ["query"],
  "additionalProperties": false
}`, t.results(), t.results())),
	}
}

// Access reports that web_search reaches the network.
func (WebSearch) Access() Access { return AccessNetwork }

// Summary shows the query.
func (WebSearch) Summary(raw json.RawMessage) string {
	var args webSearchArgs
	summaryArgs(raw, &args)
	return "Search the web for " + shorten(args.Query, 100)
}

// Run searches and lists the results.
func (t WebSearch) Run(ctx context.Context, env Env, raw json.RawMessage) Result {
	var args webSearchArgs
	if err := decodeArgs(raw, &args); err != nil {
		return Errorf("%v", err)
	}
	query := strings.TrimSpace(args.Query)
	if query == "" {
		return Errorf("query must not be empty")
	}
	count := t.results()
	if args.Count > 0 {
		count = min(args.Count, count)
	}
	ctx, cancel := context.WithTimeout(ctx, searchTimeout)
	defer cancel()
	results, err := t.Searcher.Search(ctx, query, count)
	if err != nil {
		return Errorf("searching for %q failed: %v", query, err)
	}
	if len(results) == 0 {
		return Result{Output: fmt.Sprintf("No results for %q.", query)}
	}
	return Result{Output: truncateHead(formatResults(results), outputLimit(env))}
}

func (t WebSearch) results() int {
	if t.Results <= 0 {
		return 5
	}
	return t.Results
}

func formatResults(results []websearch.Result) string {
	var b strings.Builder
	for i, r := range results {
		fmt.Fprintf(&b, "%d. %s\n   %s\n", i+1, r.Title, r.URL)
		if r.Snippet != "" {
			fmt.Fprintf(&b, "   %s\n", r.Snippet)
		}
		b.WriteByte('\n')
	}
	return strings.TrimRight(b.String(), "\n")
}
