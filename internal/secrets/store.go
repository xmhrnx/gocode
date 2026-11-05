// Package secrets stores API keys.
package secrets

import "errors"

// Store keeps named secrets.
type Store interface {
	Get(name string) (string, error)
	Set(name, value string) error
	Delete(name string) error
}

// ErrNotFound is returned for a key that was never stored.
var ErrNotFound = errors.New("secret not found")

// ProviderKeyName is the secret name of a provider's API key.
func ProviderKeyName(providerID string) string { return "provider." + providerID + ".key" }
