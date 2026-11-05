package httpapi

import (
	"encoding/json"
	"strings"
)

// errorEnvelope matches the error shapes of OpenAI, Anthropic, Gemini and
// most compatible servers: {"error":{"message":...}}, {"error":"..."} or
// {"message":"..."}.
type errorEnvelope struct {
	Error   json.RawMessage `json:"error"`
	Message string          `json:"message"`
}

// maxMessageLen keeps error messages readable when a server returns a
// whole HTML page.
const maxMessageLen = 500

// ErrorMessage extracts a human-readable message from an error response
// body, falling back to the trimmed body text and then to fallback.
func ErrorMessage(body []byte, fallback string) string {
	var envelope errorEnvelope
	if json.Unmarshal(body, &envelope) == nil {
		if message := nestedMessage(envelope.Error); message != "" {
			return Truncate(message)
		}
		if envelope.Message != "" {
			return Truncate(envelope.Message)
		}
	}
	if text := strings.TrimSpace(string(body)); text != "" {
		return Truncate(text)
	}
	return fallback
}

func nestedMessage(raw json.RawMessage) string {
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return text
	}
	var object struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(raw, &object) == nil {
		return object.Message
	}
	return ""
}

// Truncate shortens s to at most maxMessageLen bytes on a rune boundary.
func Truncate(s string) string {
	if len(s) <= maxMessageLen {
		return s
	}
	cut := maxMessageLen
	for cut > 0 && !isRuneStart(s[cut]) {
		cut--
	}
	return s[:cut] + "…"
}

func isRuneStart(b byte) bool { return b&0xC0 != 0x80 }
