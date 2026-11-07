// Package projects keeps the folder new projects go in, and makes a
// folder there for a chat whose user didn't pick one.
package projects

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// defaultName is the projects folder's name in the home folder.
const defaultName = "GoCode"

// maxNameWords and maxNameLength keep folder names short.
const (
	maxNameWords  = 5
	maxNameLength = 40
)

// Dir returns the projects folder: configured (with "~" expanded), or
// ~/GoCode when it is empty.
func Dir(configured string) string {
	home, err := os.UserHomeDir()
	configured = strings.TrimSpace(configured)
	switch {
	case configured == "" && err == nil:
		return filepath.Join(home, defaultName)
	case configured == "":
		return defaultName
	case err == nil && (configured == "~" || strings.HasPrefix(configured, "~/") || strings.HasPrefix(configured, `~\`)):
		return filepath.Join(home, configured[1:])
	}
	return configured
}

// Create makes a new folder in parent named after the start of message
// ("Build a snake game" becomes build-a-snake-game), adding -2, -3, ...
// when the name is taken, and returns its path.
func Create(parent, message string) (string, error) {
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return "", err
	}
	base := Name(message)
	for n := 1; n < 1000; n++ {
		name := base
		if n > 1 {
			name = fmt.Sprintf("%s-%d", base, n)
		}
		path := filepath.Join(parent, name)
		err := os.Mkdir(path, 0o755)
		if err == nil {
			return path, nil
		}
		if !errors.Is(err, fs.ErrExist) {
			return "", err
		}
	}
	return "", fmt.Errorf("too many folders named %s in %s", base, parent)
}

// Name turns the start of message into a folder name: lowercase words
// of letters and digits joined by hyphens, or "project".
func Name(message string) string {
	words := strings.FieldsFunc(strings.ToLower(message), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	var name string
	for i, word := range words {
		if i == maxNameWords || len(name)+len(word)+1 > maxNameLength {
			break
		}
		if name != "" {
			name += "-"
		}
		name += word
	}
	if name == "" {
		return "project"
	}
	return name
}
