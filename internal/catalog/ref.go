// Package catalog combines the built-in provider list with the user's
// provider settings and keys, and keeps the list of available models.
package catalog

import "strings"

// ModelRef names a model on a provider.
type ModelRef struct {
	Provider string
	Model    string
}

// String returns "provider/model", the form stored in settings.
func (r ModelRef) String() string {
	if r.IsZero() {
		return ""
	}
	return r.Provider + "/" + r.Model
}

// IsZero reports whether the reference names nothing.
func (r ModelRef) IsZero() bool { return r.Provider == "" || r.Model == "" }

// ParseRef parses "provider/model". Model IDs may contain slashes
// ("openrouter/anthropic/claude-x"), so only the first one separates.
func ParseRef(s string) ModelRef {
	provider, model, ok := strings.Cut(s, "/")
	if !ok {
		return ModelRef{}
	}
	return ModelRef{Provider: provider, Model: model}
}
