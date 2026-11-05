package skills

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
)

// extractZip unpacks the archive at zipPath into dest within the install
// limits. Entries that would land outside dest (zip-slip) fail the whole
// extraction; symbolic links and other special entries are skipped.
func extractZip(zipPath, dest string) error {
	archive, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("open %s: %w", zipPath, err)
	}
	defer archive.Close() // read-only, so a close error loses nothing
	if len(archive.File) > maxInstallFiles {
		return fmt.Errorf("%s has more than %d entries", zipPath, maxInstallFiles)
	}
	var limits budget
	for _, entry := range archive.File {
		if err := extractEntry(entry, dest, &limits); err != nil {
			return fmt.Errorf("extract %s: %w", zipPath, err)
		}
	}
	return nil
}

// extractEntry writes one archive entry below dest.
func extractEntry(entry *zip.File, dest string, limits *budget) error {
	name := filepath.FromSlash(entry.Name)
	if !filepath.IsLocal(name) {
		return fmt.Errorf("entry %q points outside the archive folder", entry.Name)
	}
	target := filepath.Join(dest, name)
	mode := entry.Mode()
	switch {
	case mode.IsDir():
		return os.MkdirAll(target, 0o755)
	case !mode.IsRegular():
		return nil
	}
	reader, err := entry.Open()
	if err != nil {
		return err
	}
	defer reader.Close() // read-only, so a close error loses nothing
	return limits.writeFile(target, mode, reader)
}
