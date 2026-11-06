package chat

import (
	"context"
	"fmt"
	"image"
	"time"

	"gocode/internal/attach"
)

// pasteImage attaches the image on the clipboard, if there is one; the
// image button calls it.
func (c *composer) pasteImage() {
	services := c.tab.services
	go func() {
		img, err := services.App.ReadClipboardImage(context.Background())
		switch {
		case err != nil:
			services.Error("Couldn't read the clipboard: " + err.Error())
		case img == nil:
			services.Info("The clipboard holds no image.")
		default:
			c.attachImage(img)
		}
	}()
}

// paste handles Ctrl/Cmd+V in the editor: an image on the clipboard is
// attached, anything else pastes as text.
func (c *composer) paste(pasteText func()) {
	app := c.tab.services.App
	go func() {
		img, err := app.ReadClipboardImage(context.Background())
		if err != nil || img == nil {
			app.Post(pasteText)
			return
		}
		c.attachImage(img)
	}()
}

// attachImage adds a clipboard image as an attachment; it runs off the UI
// goroutine, since encoding a large image takes a moment.
func (c *composer) attachImage(img image.Image) {
	name := fmt.Sprintf("pasted-%s.png", time.Now().Format("150405"))
	c.post(attach.FromImage(name, img))
}
