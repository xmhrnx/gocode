package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// decodeArgs unmarshals a tool's JSON arguments, wording failures so the
// model can correct its call.
func decodeArgs(raw json.RawMessage, into any) error {
	if len(bytes.TrimSpace(raw)) == 0 {
		raw = json.RawMessage("{}")
	}
	if err := json.Unmarshal(raw, into); err != nil {
		return fmt.Errorf("invalid arguments (%v); pass a JSON object matching the tool's parameter schema", err)
	}
	return nil
}

// summaryArgs decodes arguments for a one-line summary; a summary must never
// fail, so decoding errors leave into at its zero value.
func summaryArgs(raw json.RawMessage, into any) {
	_ = json.Unmarshal(raw, into)
}

// shorten cuts s to at most max runes on its first line, for summaries.
func shorten(s string, max int) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i] + " …"
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}
