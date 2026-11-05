// Package httpapi holds the HTTP plumbing shared by the provider adapters:
// sending JSON requests and turning error responses into *llm.APIError.
package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"gocode/internal/llm"
)

// maxErrorBody bounds how much of an error response is read; error bodies
// are small, and a misbehaving server shouldn't exhaust memory.
const maxErrorBody = 64 * 1024

// Send makes an HTTP request with an optional JSON body. A non-2xx response
// is consumed, closed and returned as *llm.APIError; otherwise the caller
// owns the response body.
func Send(ctx context.Context, client *http.Client, method, url string, header http.Header, body any) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encode request: %w", err)
		}
		reader = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, err
	}
	req.Header = header.Clone()
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		defer resp.Body.Close()
		data, readErr := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
		if readErr != nil && len(data) == 0 {
			return nil, &llm.APIError{StatusCode: resp.StatusCode, Message: readErr.Error()}
		}
		return nil, &llm.APIError{StatusCode: resp.StatusCode, Message: ErrorMessage(data, resp.Status)}
	}
	return resp, nil
}

// GetJSON fetches url and decodes the JSON response into out.
func GetJSON(ctx context.Context, client *http.Client, url string, header http.Header, out any) error {
	resp, err := Send(ctx, client, http.MethodGet, url, header, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

// Client returns c, or http.DefaultClient when c is nil.
func Client(c *http.Client) *http.Client {
	if c == nil {
		return http.DefaultClient
	}
	return c
}
