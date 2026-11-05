package secrets

import "errors"

// Fallback reads from the primary store and falls back to the secondary
// one when the primary is unavailable (no keyring on a minimal Linux).
type Fallback struct {
	Primary   Store
	Secondary Store
}

// Get returns the secret from the first store that has it.
func (f Fallback) Get(name string) (string, error) {
	value, err := f.Primary.Get(name)
	if err == nil {
		return value, nil
	}
	return f.Secondary.Get(name)
}

// Set stores the secret in the primary store, or the secondary one when
// the primary fails.
func (f Fallback) Set(name, value string) error {
	if err := f.Primary.Set(name, value); err == nil {
		_ = f.Secondary.Delete(name) // the primary copy now wins; a stale fallback copy only confuses
		return nil
	}
	return f.Secondary.Set(name, value)
}

// Delete removes the secret from both stores.
func (f Fallback) Delete(name string) error {
	primaryErr := f.Primary.Delete(name)
	secondaryErr := f.Secondary.Delete(name)
	if primaryErr != nil && secondaryErr != nil && !errors.Is(secondaryErr, ErrNotFound) {
		return secondaryErr
	}
	return nil
}
