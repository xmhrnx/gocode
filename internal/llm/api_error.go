package llm

import "fmt"

// APIError is an error response from a provider.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("provider returned %d: %s", e.StatusCode, e.Message)
}

// Retryable reports whether sending the same request again may succeed.
func (e *APIError) Retryable() bool {
	return e.StatusCode == 429 || e.StatusCode == 408 || e.StatusCode == 529 || e.StatusCode >= 500
}
