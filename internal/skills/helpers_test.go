package skills

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

// skillMarkdown returns a minimal valid SKILL.md.
func skillMarkdown(name, description string) string {
	return "---\nname: " + name + "\ndescription: " + description + "\n---\n# " + name + "\nDo the thing.\n"
}

// writeFiles creates files (slash-separated paths to contents) under root.
func writeFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// writeSkill creates root/folder holding a valid skill and returns its path.
func writeSkill(t *testing.T, root, folder, name, description string) string {
	t.Helper()
	writeFiles(t, root, map[string]string{folder + "/" + SkillFile: skillMarkdown(name, description)})
	return filepath.Join(root, folder)
}

// buildZip writes a zip archive of entries (names to contents) and
// returns its path; entry names are stored verbatim, unsafe ones included.
func buildZip(t *testing.T, entries map[string]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "skill.zip")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	for name, content := range entries {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

// readFile returns a file's content, failing the test if it is missing.
func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
