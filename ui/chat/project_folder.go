package chat

import (
	"gocode/internal/projects"
)

// ensureFolder gives a chat without a folder a new one in the projects
// folder, named after its first message, and reports whether the chat
// has a folder now.
func (c *ChatTab) ensureFolder(firstMessage string) bool {
	if c.session.WorkDir != "" {
		return true
	}
	dir, err := projects.Create(c.projectsDir(), firstMessage)
	if err != nil {
		c.services.Error("Couldn't create a folder for this chat: " + err.Error())
		return false
	}
	c.session.WorkDir = dir
	return true
}

// projectsDir is where new project folders go.
func (c *ChatTab) projectsDir() string {
	return projects.Dir(c.services.Settings.General.ProjectsDir)
}

// folderLabel names the chat's folder, or "Default" while it has none:
// it will get one in the projects folder (Settings → General).
func (c *ChatTab) folderLabel() string {
	if c.session.WorkDir == "" {
		return "Default"
	}
	return shortPath(c.session.WorkDir)
}
