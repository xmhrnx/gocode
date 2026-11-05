package chat

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/xmhrnx/giode"

	"gocode/internal/attach"
)

// pickFiles shows the file picker and attaches what the user chooses.
func (c *composer) pickFiles() {
	app, limits := c.tab.services.App, c.limits()
	go func() {
		files, err := app.PickFiles(context.Background(), giode.PickOptions{Multiple: true})
		if errors.Is(err, giode.ErrCanceled) {
			return
		}
		if err != nil {
			c.tab.services.Error("Couldn't open the file picker: " + err.Error())
			return
		}
		for _, f := range files {
			att, err := attach.Read(f.Name, f, limits)
			f.Close()
			c.post(att, err)
		}
	}()
}

// attachPaths attaches files dropped on the window.
func (c *composer) attachPaths(paths []string) {
	limits := c.limits()
	go func() {
		for _, path := range paths {
			c.post(readPath(path, limits))
		}
	}()
}

func readPath(path string, limits attach.Limits) (attach.Attachment, error) {
	f, err := os.Open(path)
	if err != nil {
		return attach.Attachment{}, err
	}
	defer f.Close()
	return attach.Read(filepath.Base(path), f, limits)
}

// post hands a read attachment, or the reason it failed, to the UI.
func (c *composer) post(att attach.Attachment, err error) {
	if err != nil {
		c.tab.services.Error(err.Error())
		return
	}
	c.tab.services.App.Post(func() { c.add(att) })
}

func (c *composer) limits() attach.Limits {
	chat := c.tab.services.Settings.Chat
	return attach.Limits{
		MaxBytes:     int64(chat.MaxAttachmentMB) << 20,
		MaxTextBytes: int64(chat.InlineTextFilesKB) << 10,
	}
}
