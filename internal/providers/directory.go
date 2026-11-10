package providers

import (
	_ "embed"
	"encoding/json"
	"sync"
)

// directoryFile holds providers from the models.dev directory that speak
// the OpenAI-compatible or Anthropic API, beyond the ones defined in code.
//
//go:embed directory.json
var directoryFile []byte

// directory is the parsed directory.json.
type directory struct {
	Source    string       `json:"source"`
	Providers []Definition `json:"providers"`
}

var loadDirectory = sync.OnceValue(func() []Definition {
	var d directory
	if err := json.Unmarshal(directoryFile, &d); err != nil {
		panic("providers: directory.json: " + err.Error()) // embedded, so a test catches it
	}
	return d.Providers
})

// directoryProviders returns a copy of the directory's providers.
func directoryProviders() []Definition {
	return append([]Definition(nil), loadDirectory()...)
}
