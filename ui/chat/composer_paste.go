package chat

import (
	"context"
	"fmt"
	"time"

	"gocode/internal/attach"
)

// pasteImage attaches the image on the clipboard, if there is one.
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
			name := fmt.Sprintf("pasted-%s.png", time.Now().Format("150405"))
			c.post(attach.FromImage(name, img))
		}
	}()
}
