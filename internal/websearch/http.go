package websearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// maxErrorBody bounds how much of an error response is quoted.
const maxErrorBody = 300

// doJSON sends a request (a GET when body is nil) and decodes the JSON
// answer into out.
func doJSON(ctx context.Context, client *http.Client, method, url string, headers map[string]string, body, out any) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for name, value := range headers {
		req.Header.Set(name, value)
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
		return fmt.Errorf("search failed: %s: %s", resp.Status, strings.TrimSpace(string(data)))
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode search results: %w", err)
	}
	return nil
}

var tags = regexp.MustCompile(`<[^>]*>`)

// plain removes markup (Brave marks matches with <strong>) and squeezes
// whitespace.
func plain(text string) string {
	text = tags.ReplaceAllString(text, "")
	for _, entity := range [][2]string{{"&amp;", "&"}, {"&lt;", "<"}, {"&gt;", ">"}, {"&quot;", `"`}, {"&#39;", "'"}} {
		text = strings.ReplaceAll(text, entity[0], entity[1])
	}
	return strings.Join(strings.Fields(text), " ")
}
