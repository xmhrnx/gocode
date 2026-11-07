package settingsview

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"

	"gocode/internal/skills"
	"gocode/ui/core"
	"gocode/ui/theme"

	"gocode/ui/components"
)

// skillsPage lists installed skills, switches them on and off, and
// installs new ones from a folder, a .zip or a URL.
type skillsPage struct {
	services                     *core.Services
	found                        []skills.Skill
	problems                     []skills.Problem
	rows                         map[string]*skillRow
	url                          *input.Input
	fromFolder, fromZip, fromURL *giode.ButtonElement
	busy                         bool
}

func skillEntries(s *core.Services) []entry {
	return []entry{{
		title:       "Skills",
		description: "Skills are folders with a SKILL.md: instructions, and sometimes scripts, for a kind of task. The agent sees each skill's name and description and loads a skill when a task matches. Skills in a project's .gocode/skills or .claude/skills folder are found too.",
		keywords:    "skills skill install plugin instructions SKILL.md zip github",
		wide:        true,
		control:     newSkillsPage(s),
	}}
}

func newSkillsPage(s *core.Services) *skillsPage {
	p := &skillsPage{services: s, rows: map[string]*skillRow{}}
	p.url = giode.Input.Text("https://github.com/owner/repo or a .zip URL", fieldStyle(-1)).Submit(func(string) { p.installURL() })
	components.AddEditMenu(s.App, p.url, components.EditOptions{})
	p.fromFolder = smallButton("Install from folder…", false).OnClick(p.installFolder)
	p.fromZip = smallButton("Install from .zip…", false).OnClick(p.installZip)
	p.fromURL = smallButton("Install", true).OnClick(p.installURL)
	p.refresh()
	return p
}

// refresh rediscovers the installed skills off the UI goroutine.
func (p *skillsPage) refresh() {
	userDir, workDir := p.services.SkillsDir, "" // project skills show in each chat
	go func() {
		found, problems := skills.Discover(userDir, workDir)
		p.services.App.Post(func() { p.found, p.problems = found, problems })
	}()
}

func (p *skillsPage) Layout(gtx giode.Context) giode.Dimensions {
	children := []giode.Element{
		giode.Row(giode.Styles{Gap: 8}, p.fromFolder, p.fromZip),
		giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{FlexGrow: 1}, p.url), p.fromURL),
		groupTitle(fmt.Sprintf("Installed (%d)", len(p.found))),
	}
	for _, skill := range p.found {
		children = append(children, p.row(skill))
	}
	if len(p.found) == 0 {
		children = append(children, giode.Text("No skills installed yet.", theme.Caption()))
	}
	for _, problem := range p.problems {
		children = append(children, giode.Text(problem.Dir+": "+problem.Err.Error(),
			giode.Styles{Color: theme.Danger, FontSize: theme.Sizes.Small}))
	}
	children = append(children, giode.Text("Installed in "+p.services.SkillsDir, theme.Caption()))
	return giode.Stack(giode.Styles{Width: -1, Gap: 10}, children...).Layout(gtx)
}

func (p *skillsPage) row(skill skills.Skill) *skillRow {
	key := string(skill.Source) + "/" + skill.Name
	row, ok := p.rows[key]
	if !ok {
		row = newSkillRow(p, skill)
		p.rows[key] = row
	}
	return row
}

// install runs an installation in the background and reports the result.
func (p *skillsPage) install(run func(ctx context.Context) ([]skills.Skill, error)) {
	if p.busy {
		return
	}
	p.busy = true
	go func() {
		installed, err := run(context.Background())
		p.services.App.Post(func() {
			p.busy = false
			p.report(installed, err)
			p.refresh()
		})
	}()
}

func (p *skillsPage) report(installed []skills.Skill, err error) {
	names := make([]string, len(installed))
	for i, skill := range installed {
		names[i] = skill.Name
	}
	switch {
	case errors.Is(err, skills.ErrExists):
		p.services.Error("That skill is already installed. Remove it first to replace it.")
	case err != nil:
		p.services.Error("Couldn't install: " + err.Error())
	case len(names) > 0:
		p.services.Info("Installed " + strings.Join(names, ", "))
	}
}
