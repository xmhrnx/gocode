package settingsview

import (
	"context"
	"errors"
	"strings"

	"github.com/xmhrnx/giode"

	"gocode/internal/skills"
)

// installFolder installs a skill folder the user picks.
func (p *skillsPage) installFolder() {
	p.install(func(ctx context.Context) ([]skills.Skill, error) {
		dir, err := p.services.App.PickFolder(ctx, giode.FolderOptions{})
		if err != nil || dir == "" {
			return nil, ignoreCanceled(err)
		}
		skill, err := skills.InstallFromFolder(dir, p.services.SkillsDir, false)
		return []skills.Skill{skill}, err
	})
}

// installZip installs a skill from a .zip file the user picks.
func (p *skillsPage) installZip() {
	p.install(func(ctx context.Context) ([]skills.Skill, error) {
		files, err := p.services.App.PickFiles(ctx, giode.PickOptions{Types: []string{"zip"}})
		if err != nil || len(files) == 0 {
			return nil, ignoreCanceled(err)
		}
		files[0].Close() // only the path is needed; the archive is read by path
		skill, err := skills.InstallFromZip(files[0].Path, p.services.SkillsDir, false)
		return []skills.Skill{skill}, err
	})
}

// installURL installs from the URL typed in the field.
func (p *skillsPage) installURL() {
	url := strings.TrimSpace(p.url.Text())
	if url == "" {
		return
	}
	p.url.SetText("")
	p.install(func(ctx context.Context) ([]skills.Skill, error) {
		return skills.InstallFromURL(ctx, p.services.HTTP, url, p.services.SkillsDir, false)
	})
}

func ignoreCanceled(err error) error {
	if errors.Is(err, giode.ErrCanceled) {
		return nil
	}
	return err
}
