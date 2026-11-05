package secrets

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sync"

	"gocode/internal/settings"
)

// FileStore keeps secrets in a JSON file readable only by the user. It is
// the fallback for systems without a credential store.
type FileStore struct {
	Path string
	mu   sync.Mutex
}

// Get returns the named secret.
func (f *FileStore) Get(name string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	all, err := f.read()
	if err != nil {
		return "", err
	}
	value, ok := all[name]
	if !ok {
		return "", ErrNotFound
	}
	return value, nil
}

// Set stores the named secret.
func (f *FileStore) Set(name, value string) error {
	return f.update(func(all map[string]string) { all[name] = value })
}

// Delete removes the named secret.
func (f *FileStore) Delete(name string) error {
	return f.update(func(all map[string]string) { delete(all, name) })
}

func (f *FileStore) update(change func(map[string]string)) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	all, err := f.read()
	if err != nil {
		return err
	}
	change(all)
	data, err := json.Marshal(all)
	if err != nil {
		return fmt.Errorf("encode secrets: %w", err)
	}
	if err := settings.WriteFileAtomic(f.Path, data); err != nil {
		return err
	}
	return os.Chmod(f.Path, 0o600)
}

func (f *FileStore) read() (map[string]string, error) {
	all := map[string]string{}
	data, err := os.ReadFile(f.Path)
	if errors.Is(err, fs.ErrNotExist) {
		return all, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read secrets: %w", err)
	}
	if err := json.Unmarshal(data, &all); err != nil {
		return nil, fmt.Errorf("parse secrets: %w", err)
	}
	return all, nil
}
