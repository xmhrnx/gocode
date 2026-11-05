package skills

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
)

// errNoSkill is returned when a download or archive holds no SKILL.md.
var errNoSkill = errors.New("no folder containing " + SkillFile + " was found")

// findSkillDirs returns the folders under root, at most maxDepth levels
// down, that contain SKILL.md. Folders inside a skill are not searched, and
// hidden folders (.git) and macOS archive metadata are skipped.
func findSkillDirs(root string, maxDepth int) ([]string, error) {
	root = filepath.Clean(root) // so depth can strip it as a prefix
	var dirs []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			return nil
		}
		if path != root && isIgnoredFolder(entry.Name()) {
			return fs.SkipDir
		}
		if hasSkillFile(path) {
			dirs = append(dirs, path)
			return fs.SkipDir
		}
		if depth(root, path) >= maxDepth {
			return fs.SkipDir
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(dirs) == 0 {
		return nil, errNoSkill
	}
	return dirs, nil
}

func isIgnoredFolder(name string) bool {
	return strings.HasPrefix(name, ".") || name == "__MACOSX"
}

// depth counts the folders between root and path, a folder WalkDir
// reached from root.
func depth(root, path string) int {
	return strings.Count(strings.TrimPrefix(path, root), string(filepath.Separator))
}
