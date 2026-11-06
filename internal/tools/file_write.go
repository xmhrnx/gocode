package tools

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gocode/internal/fsretry"
)

// existingFile is what a write needs to know about the file it replaces.
type existingFile struct {
	content string
	mode    fs.FileMode
	exists  bool
}

// loadExisting reads the file at abs, if any. Directories and binary files
// are refused since the text tools would corrupt them.
func loadExisting(abs string) (existingFile, error) {
	info, err := os.Stat(abs)
	if os.IsNotExist(err) {
		return existingFile{mode: 0o644}, nil
	}
	if err != nil {
		return existingFile{}, err
	}
	if info.IsDir() {
		return existingFile{}, &fs.PathError{Op: "write", Path: abs, Err: errIsDirectory}
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return existingFile{}, err
	}
	if looksBinary(data) {
		return existingFile{}, &fs.PathError{Op: "edit", Path: abs, Err: errBinaryFile}
	}
	return existingFile{content: string(data), mode: info.Mode().Perm(), exists: true}, nil
}

// writeText writes content to abs with mode, creating parent directories.
func writeText(abs, content string, mode fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}
	if err := fsretry.Do(func() error { return os.WriteFile(abs, []byte(content), mode) }); err != nil {
		return explainWriteError(err)
	}
	// WriteFile only applies mode to new files; an existing file keeps its
	// own, so set it explicitly in case the umask or a prior chmod differs.
	return os.Chmod(abs, mode)
}

// fileDiff renders the change to the file at abs as a unified diff with
// paths shown relative to the project root.
func fileDiff(env Env, abs string, before existingFile, after string) string {
	name := displayPath(env, abs)
	oldName := "a/" + name
	if !before.exists {
		oldName = "/dev/null"
	}
	return unifiedDiff(oldName, "b/"+name, before.content, after)
}

// explainWriteError adds what the model can do about a refused write, so
// it doesn't retry the same way or fall back to writing through the shell.
func explainWriteError(err error) error {
	if !errors.Is(err, fs.ErrPermission) {
		return err
	}
	return fmt.Errorf("%w. The file may be read-only, open in a program that locks it, or in a folder the system "+
		"protects (Windows' Controlled Folder Access blocks unknown apps from Documents and Desktop). "+
		"Tell the user; writing through the shell will fail the same way", err)
}
