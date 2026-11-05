package chat

import (
	"slices"

	"gocode/internal/skills"
	"gocode/internal/tools"
)

// toolRegistry returns the tools for a run: the built-in ones, the skill
// loader, and the tools of the connected MCP servers.
func (c *ChatTab) toolRegistry() *tools.Registry {
	all := append(tools.Builtin(), skills.NewTool(c.skillLister()))
	all = append(all, c.services.MCP.Tools()...)
	return tools.NewRegistry(all...)
}

// skillLister returns a function listing the enabled skills for this
// chat's folder. It reads the disk, so it runs on the run's goroutine.
func (c *ChatTab) skillLister() func() []skills.Skill {
	userDir, workDir := c.services.SkillsDir, c.session.WorkDir
	disabled := slices.Clone(c.services.Settings.Skills.Disabled)
	return func() []skills.Skill {
		found, _ := skills.Discover(userDir, workDir) // broken skills are shown in Settings → Skills
		return skills.Enabled(found, disabled)
	}
}
