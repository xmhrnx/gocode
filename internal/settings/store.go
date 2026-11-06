package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

// Load reads settings from path. A missing file yields the defaults.
func Load(path string) (Settings, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return Default(), nil
	}
	if err != nil {
		return Default(), fmt.Errorf("read settings: %w", err)
	}
	loaded := Default()
	if err := json.Unmarshal(data, &loaded); err != nil {
		return Default(), fmt.Errorf("parse settings %s: %w", path, err)
	}
	loaded.migrate()
	loaded.fillDefaults()
	return loaded, nil
}

// Save writes settings to path atomically, so a crash mid-write never
// leaves a truncated file.
func Save(path string, s Settings) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("encode settings: %w", err)
	}
	return WriteFileAtomic(path, data)
}
