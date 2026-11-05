package openai

import (
	"encoding/json"
	"net/http"

	"gocode/internal/llm"
	"gocode/internal/llm/httpapi"
)

// streamError converts an error object sent as a stream payload (as
// OpenRouter and some servers do after a 200 response) into *llm.APIError.
// data is the whole payload; errorField is its "error" member.
func streamError(data []byte, errorField json.RawMessage) error {
	var detail struct {
		Code json.RawMessage `json:"code"`
	}
	status := http.StatusInternalServerError // mid-stream failures are server-side unless a code says otherwise
	if json.Unmarshal(errorField, &detail) == nil {
		var code int
		if json.Unmarshal(detail.Code, &code) == nil && code >= 400 && code < 600 {
			status = code
		}
	}
	return &llm.APIError{StatusCode: status, Message: httpapi.ErrorMessage(data, "stream error")}
}
