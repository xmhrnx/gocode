package skills

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
)

// maxListedFiles bounds the file listing so a skill bundling a large tree
// cannot flood the model's context.
const maxListedFiles = 100

// Load returns skill's instructions (the SKILL.md body without frontmatter)
// and the other files in its folder as slash-separated relative paths,
// at most 100 of them, skipping hidden files and folders.
func Load(skill Skill) (body string, files []string, err error) {
	_, body, err = readSkill(skill.Dir, skill.Source)
	if err != nil {
		return "", nil, err
	}
	files, err = listFiles(skill.Dir)
	if err != nil {
		return "", nil, err
	}
	return strings.TrimSpace(body), files, nil
}

// listFiles walks dir in lexical order and collects its visible files.
func listFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == dir {
			return nil
		}
		if strings.HasPrefix(entry.Name(), ".") {
			return skipEntry(entry)
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		if entry.IsDir() || rel == SkillFile {
			return nil
		}
		if len(files) == maxListedFiles {
			return fs.SkipAll
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if err != nil && !errors.Is(err, fs.SkipAll) {
		return nil, err
	}
	return files, nil
}

// skipEntry skips a whole folder, or just the one file.
func skipEntry(entry fs.DirEntry) error {
	if entry.IsDir() {
		return fs.SkipDir
	}
	return nil
}
