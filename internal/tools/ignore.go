package tools

import (
	"path/filepath"
	"strings"
)

// ignoreRules decides which paths walking tools skip: the built-in
// directory names plus the patterns of the project's root .gitignore.
type ignoreRules struct {
	root     string
	patterns []gitignorePattern
}

// ignoredDirNames are skipped by every walking tool: they are large, rarely
// what the model is looking for, and not hidden by every project's
// .gitignore.
var ignoredDirNames = map[string]bool{
	".git":         true,
	"node_modules": true,
	".venv":        true,
	"venv":         true,
	"__pycache__":  true,
	".idea":        true,
	".gradle":      true,
	".next":        true,
	".cache":       true,
}

// loadIgnoreRules reads root/.gitignore. An unreadable .gitignore only
// loses its patterns, since searching must still work without it.
func loadIgnoreRules(root string) *ignoreRules {
	patterns, err := parseGitignore(filepath.Join(root, ".gitignore"))
	if err != nil {
		patterns = nil
	}
	return &ignoreRules{root: root, patterns: patterns}
}

// skip reports whether the absolute path should be left out of a walk.
func (r *ignoreRules) skip(abs string, isDir bool) bool {
	if isDir && ignoredDirNames[filepath.Base(abs)] {
		return true
	}
	rel, err := filepath.Rel(r.root, abs)
	if err != nil {
		return false
	}
	rel = filepath.ToSlash(rel)
	if rel == "." || rel == ".." || strings.HasPrefix(rel, "../") {
		return false
	}
	for _, p := range r.patterns {
		if p.matches(rel, isDir) {
			return true
		}
	}
	return false
}
