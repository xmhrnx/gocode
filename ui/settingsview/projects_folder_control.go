package settingsview

import (
	"context"
	"errors"
	"strings"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"

	"gocode/internal/projects"
	"gocode/ui/components"
	"gocode/ui/core"
)

// projectsFolderControl edits the projects folder: its path, shown in
// full even while it is the default, and a folder picker.
type projectsFolderControl struct {
	services *core.Services
	field    *input.Input
	browse   *giode.ButtonElement
}

func newProjectsFolderControl(s *core.Services) *projectsFolderControl {
	c := &projectsFolderControl{services: s}
	c.field = giode.Input.Text(projects.Display(projects.Dir("")), fieldStyle(260))
	c.field.SetText(projects.Display(projects.Dir(s.Settings.General.ProjectsDir)))
	c.field.OnChange(c.store)
	components.AddEditMenu(s.App, c.field, components.EditOptions{})
	c.browse = smallButton("Browse…", false).OnClick(c.pick)
	return c
}

func (c *projectsFolderControl) Layout(gtx giode.Context) giode.Dimensions {
	return giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter}, c.field, c.browse).Layout(gtx)
}

// store keeps the default as empty, so the setting follows the default.
func (c *projectsFolderControl) store(text string) {
	text = strings.TrimSpace(text)
	if text == projects.Display(projects.Dir("")) || projects.Dir(text) == projects.Dir("") {
		text = ""
	}
	c.services.Settings.General.ProjectsDir = text
	c.services.SettingsChanged()
}

func (c *projectsFolderControl) pick() {
	s := c.services
	go func() {
		dir, err := s.App.PickFolder(context.Background(), giode.FolderOptions{Title: "Projects folder"})
		if errors.Is(err, giode.ErrCanceled) || dir == "" {
			return
		}
		if err != nil {
			s.Error("Couldn't open the folder picker: " + err.Error())
			return
		}
		s.App.Post(func() {
			c.field.SetText(projects.Display(dir))
			c.store(dir)
			giode.InvalidateCaches() // the field changed from code
		})
	}()
}
