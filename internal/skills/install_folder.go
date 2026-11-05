package skills

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// InstallFromFolder copies the skill folder src (which must hold a valid
// SKILL.md) into userDir/<name>. An installed skill of the same name is
// replaced only when overwrite is true; otherwise ErrExists is returned.
func InstallFromFolder(src, userDir string, overwrite bool) (Skill, error) {
	skill, _, err := readSkill(src, SourceUser)
	if err != nil {
		return Skill{}, err
	}
	if err := checkNotNested(src, userDir); err != nil {
		return Skill{}, err
	}
	dest := filepath.Join(userDir, skill.Name)
	if err := checkDestination(dest, overwrite); err != nil {
		return Skill{}, err
	}
	if err := replaceWithCopy(src, userDir, dest); err != nil {
		return Skill{}, fmt.Errorf("install skill %s: %w", skill.Name, err)
	}
	skill.Dir = dest
	return skill, nil
}

// checkDestination refuses to replace an installed skill unless asked.
func checkDestination(dest string, overwrite bool) error {
	_, err := os.Lstat(dest)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil
	case err != nil:
		return err
	case !overwrite:
		return fmt.Errorf("%w: %s", ErrExists, dest)
	}
	return nil
}

// checkNotNested refuses a source that contains the skills directory,
// which would make the copy include itself.
func checkNotNested(src, userDir string) error {
	absSrc, err := filepath.Abs(src)
	if err != nil {
		return err
	}
	absUserDir, err := filepath.Abs(userDir)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(absSrc, absUserDir)
	if err == nil && (rel == "." || filepath.IsLocal(rel)) {
		return fmt.Errorf("cannot install %s: it contains the skills directory", src)
	}
	return nil
}

// replaceWithCopy copies src into a hidden staging folder next to dest and
// then swaps it in, so a failed copy never leaves a half-installed skill.
// Staging first also makes reinstalling a skill from its own folder safe.
func replaceWithCopy(src, userDir, dest string) error {
	if err := os.MkdirAll(userDir, 0o755); err != nil {
		return err
	}
	staging, err := os.MkdirTemp(userDir, ".installing-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(staging) // cleans up after a failure; gone after success
	if err := copyDir(src, staging); err != nil {
		return err
	}
	if err := removeSkillDir(dest); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return os.Rename(staging, dest)
}
