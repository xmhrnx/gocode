package chat

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/xmhrnx/giode"

	"gocode/internal/llm"
	"gocode/internal/session"
)

// openMedia writes media the model made to GoCode's cache folder, once,
// and opens it in the app the system picks for its type.
func (c *ChatTab) openMedia(it item, media llm.Media) {
	services, dir := c.services, mediaCacheDir(c.session.ID, it)
	go func() {
		path, err := writeOnce(filepath.Join(dir, filepath.Base(media.Name)), media.Data)
		if err == nil {
			err = services.App.OpenWith(context.Background(), path)
		}
		if err != nil && !errors.Is(err, giode.ErrCanceled) {
			services.Error("Couldn't open " + media.Name + ": " + err.Error())
		}
	}()
}

// mediaCacheDir is where one reply part's media is kept for opening:
// a folder per chat and part, so files with the same name don't clash.
func mediaCacheDir(sessionID string, it item) string {
	dir := session.MediaCacheDir(sessionID)
	if dir == "" {
		dir = filepath.Join(os.TempDir(), "GoCode", "media", sessionID)
	}
	return filepath.Join(dir, it.key)
}

// writeOnce writes data to path unless a file of the same size is
// already there, and returns the path.
func writeOnce(path string, data []byte) (string, error) {
	if st, err := os.Stat(path); err == nil && st.Size() == int64(len(data)) {
		return path, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	return path, os.WriteFile(path, data, 0o644)
}
