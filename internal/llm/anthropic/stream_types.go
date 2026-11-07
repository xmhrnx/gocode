package anthropic

import (
	"encoding/json"
	"net/http"

	"gocode/internal/llm"
)

// streamEvent is the union of all streamed event payloads; which fields
// are set depends on Type.
type streamEvent struct {
	Type         string          `json:"type"`
	Index        int             `json:"index"`
	Message      wireStart       `json:"message"`
	ContentBlock json.RawMessage `json:"content_block"`
	Delta        wireDelta       `json:"delta"`
	Usage        *wireUsage      `json:"usage"`
	Error        wireError       `json:"error"`
}

type wireStart struct {
	Usage *wireUsage `json:"usage"`
}

// wireBlock is a content_block_start block.
type wireBlock struct {
	Type     string `json:"type"`
	Text     string `json:"text"`
	Thinking string `json:"thinking"`
	Data     string `json:"data"`
	ID       string `json:"id"`
	Name     string `json:"name"`
}

// wireDelta is a content_block_delta or message_delta delta.
type wireDelta struct {
	Type        string `json:"type"`
	Text        string `json:"text"`
	Thinking    string `json:"thinking"`
	Signature   string `json:"signature"`
	PartialJSON string `json:"partial_json"`
	StopReason  string `json:"stop_reason"`
}

type wireUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
}

type wireError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// errorStatus maps error types to the HTTP status the API uses for them,
// so a mid-stream error is retried exactly like the same error up front.
var errorStatus = map[string]int{
	"invalid_request_error": http.StatusBadRequest,
	"authentication_error":  http.StatusUnauthorized,
	"permission_error":      http.StatusForbidden,
	"not_found_error":       http.StatusNotFound,
	"request_too_large":     http.StatusRequestEntityTooLarge,
	"rate_limit_error":      http.StatusTooManyRequests,
	"api_error":             http.StatusInternalServerError,
	"overloaded_error":      529,
}

func (e wireError) apiError() *llm.APIError {
	status, ok := errorStatus[e.Type]
	if !ok {
		status = http.StatusInternalServerError
	}
	message := e.Message
	if message == "" {
		message = e.Type
	}
	return &llm.APIError{StatusCode: status, Message: message}
}
