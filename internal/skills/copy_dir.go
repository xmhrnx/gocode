package skills

import (
	"io/fs"
	"os"
	"path/filepath"
)

// copyDir copies the regular files and folders of src into dst within the
// install limits. Symbolic links are skipped because they could point
// outside the skill, and .git is skipped because it is clone metadata.
func copyDir(src, dst string) error {
	var limits budget
	return filepath.WalkDir(src, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		switch {
		case entry.IsDir() && entry.Name() == ".git":
			return fs.SkipDir
		case entry.IsDir():
			return os.MkdirAll(target, 0o755)
		case entry.Type().IsRegular():
			return copyFile(path, target, &limits)
		default:
			return nil // symlinks, devices and other special files
		}
	})
}

// copyFile copies one regular file into target.
func copyFile(path, target string, limits *budget) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close() // read-only, so a close error loses nothing
	return limits.writeFile(target, info.Mode(), file)
}
