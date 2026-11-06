package tools

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// resolvePath turns a path given by the model into a clean absolute path.
// Relative paths resolve against the project root and "~/" against the
// user's home directory.
func resolvePath(env Env, path string) (string, error) {
	path = nativePath(strings.TrimSpace(path))
	if path == "" {
		return "", errors.New("path is required")
	}
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[1:])
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(env.WorkDir, path)
	}
	return filepath.Clean(path), nil
}

// displayPath shows abs relative to the project root when it lies inside
// it, so results stay short and can be passed back to other tools as-is.
func displayPath(env Env, abs string) string {
	if env.WorkDir == "" {
		return abs
	}
	rel, err := filepath.Rel(env.WorkDir, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return abs
	}
	return filepath.ToSlash(rel)
}
