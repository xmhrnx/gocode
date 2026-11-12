package chat

import (
	"context"
	"errors"

	"github.com/xmhrnx/giode"

	"gocode/internal/llm"
)

// saveMedia asks where to save media the model made and writes it there.
func (c *ChatTab) saveMedia(media llm.Media) {
	services := c.services
	go func() {
		file, err := services.App.SaveFile(context.Background(), media.Name)
		if errors.Is(err, giode.ErrCanceled) {
			return
		}
		if err == nil {
			_, err = file.Write(media.Data)
			if closeErr := file.Close(); err == nil {
				err = closeErr
			}
		}
		if err != nil {
			services.Error("Couldn't save " + media.Name + ": " + err.Error())
			return
		}
		services.Info("Saved " + media.Name)
	}()
}
