package catalog

import (
	"errors"
	"fmt"
	"net/http"

	"gocode/internal/llm"
	"gocode/internal/providers"
	"gocode/internal/secrets"
)

// NewClient returns a client for the provider using its stored API key.
// Reading the key may block on the system keyring, so call it off the UI
// goroutine.
func NewClient(def providers.Definition, keys secrets.Store, httpClient *http.Client) (llm.Client, error) {
	key, err := keys.Get(secrets.ProviderKeyName(def.ID))
	switch {
	case errors.Is(err, secrets.ErrNotFound) && def.KeyOptional:
		key = ""
	case errors.Is(err, secrets.ErrNotFound):
		return nil, fmt.Errorf("%s needs an API key: add it in Settings → Providers", def.Name)
	case err != nil:
		return nil, fmt.Errorf("read the %s API key: %w", def.Name, err)
	}
	return providers.NewClient(def, key, httpClient)
}
