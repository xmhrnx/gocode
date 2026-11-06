package settings

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gocode/internal/fsretry"
)

// pathLocks lets one write at a time reach each file: saves run in the
// background, and on Windows two renames onto the same file at once make
// one of them fail with "Access is denied".
var pathLocks sync.Map // path -> *sync.Mutex

// WriteFileAtomic writes data to a temporary file beside path and renames
// it into place, so a crash mid-write never leaves a truncated file.
func WriteFileAtomic(path string, data []byte) error {
	lock, _ := pathLocks.LoadOrStore(filepath.Clean(path), &sync.Mutex{})
	lock.(*sync.Mutex).Lock()
	defer lock.(*sync.Mutex).Unlock()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create %s: %w", filepath.Dir(path), err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write %s: %w", path, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close %s: %w", path, err)
	}
	return fsretry.Do(func() error { return os.Rename(tmp.Name(), path) })
}
