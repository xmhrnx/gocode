package session

import (
	"os"
	"path/filepath"
)

// MediaCacheDir is where copies of a chat's generated media are kept for
// opening in other apps, or "" without a cache folder.
func MediaCacheDir(id string) string {
	base, err := os.UserCacheDir()
	if err != nil || id == "" {
		return ""
	}
	return filepath.Join(base, "GoCode", "media", id)
}
