package skills

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Remove deletes the user skill userDir/<name>. Project skills belong to
// the project's repository and are never removed here.
func Remove(userDir, name string) error {
	// Validating first keeps a name like "../x" from escaping userDir.
	if err := validateName(name); err != nil {
		return err
	}
	err := removeSkillDir(filepath.Join(userDir, name))
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("skill %q is not installed in %s", name, userDir)
	}
	return err
}

// removeSkillDir deletes a skill folder; a symlinked skill loses only the
// link, never the folder it points to.
func removeSkillDir(dir string) error {
	info, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	if info.Mode()&fs.ModeSymlink != 0 {
		return os.Remove(dir)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a skill folder", dir)
	}
	return os.RemoveAll(dir)
}
