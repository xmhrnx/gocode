package settings

import (
	"fmt"
	"os"
	"path/filepath"
)

// Dir returns the folder GoCode keeps its files in, creating it if needed.
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("find config folder: %w", err)
	}
	dir := filepath.Join(base, "gocode")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("create %s: %w", dir, err)
	}
	return dir, nil
}
