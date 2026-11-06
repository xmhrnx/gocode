package chat

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/xmhrnx/giode"

	"gocode/ui/components"
)

// header is the bar above the transcript showing the project folder,
// which the user can change before the chat starts.
type header struct {
	tab     *ChatTab
	folder  *components.PillButton
	compact *components.IconButton
}

func newHeader(tab *ChatTab) *header {
	h := &header{tab: tab}
	h.folder = components.NewPillButton("folder", "Change the project folder", h.pickFolder)
	h.compact = components.NewIconButton("layers", "Compact the conversation: summarize it so the model has room again", tab.compactNow).SetSize(15)
	return h
}

func (h *header) Layout(gtx giode.Context) giode.Dimensions {
	h.folder.SetLabel(shortPath(h.tab.session.WorkDir))
	return giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
		Padding: giode.Inset{Left: 12, Right: 12, Top: 6, Bottom: 2}},
		h.folder,
		giode.Box(giode.Styles{FlexGrow: 1}),
		contextLabel(h.tab),
		usageLabel(h.tab),
		h.compact,
	).Layout(gtx)
}

func (h *header) pickFolder() {
	services := h.tab.services
	go func() {
		dir, err := services.App.PickFolder(context.Background(), giode.FolderOptions{})
		if errors.Is(err, giode.ErrCanceled) || dir == "" {
			return
		}
		if err != nil {
			services.Error("Couldn't open the folder picker: " + err.Error())
			return
		}
		services.App.Post(func() {
			h.tab.session.WorkDir = dir
			h.tab.changed()
		})
	}()
}

// shortPath shows a path with the home folder as "~".
func shortPath(path string) string {
	home, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(path, home) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return filepath.Clean(path)
}
