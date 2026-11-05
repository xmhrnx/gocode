// Package platform adapts operating system services to GoCode's needs.
package platform

import (
	"errors"

	"github.com/xmhrnx/giode"

	"gocode/internal/secrets"
)

// Keychain stores secrets in the system credential store: the macOS
// Keychain, Windows DPAPI or the Linux Secret Service.
type Keychain struct{}

// NewKeychain returns the system store for GoCode's secrets.
func NewKeychain() Keychain {
	giode.SecureStore.Service = "dev.gocode.app"
	return Keychain{}
}

// Get returns the named secret.
func (Keychain) Get(name string) (string, error) {
	value, err := giode.SecureStore.Get(name)
	if errors.Is(err, giode.ErrNotFound) {
		return "", secrets.ErrNotFound
	}
	return string(value), err
}

// Set stores the named secret.
func (Keychain) Set(name, value string) error {
	return giode.SecureStore.Set(name, []byte(value))
}

// Delete removes the named secret.
func (Keychain) Delete(name string) error {
	err := giode.SecureStore.Delete(name)
	if errors.Is(err, giode.ErrNotFound) {
		return nil
	}
	return err
}
