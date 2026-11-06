package settingsview

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/toggle"

	"gocode/internal/skills"
	"gocode/ui/components"
	"gocode/ui/theme"
)

// skillRow shows one skill with a switch and, for user skills, a remove
// button.
type skillRow struct {
	skill   skills.Skill
	enabled *toggle.Switch
	remove  *components.IconButton
}

func newSkillRow(p *skillsPage, skill skills.Skill) *skillRow {
	s := p.services
	r := &skillRow{skill: skill}
	r.enabled = giode.Input.Switch("")
	r.enabled.SetChecked(!contains(s.Settings.Skills.Disabled, skill.Name))
	r.enabled.OnChange(func(on bool) {
		disabled := remove(s.Settings.Skills.Disabled, skill.Name)
		if !on {
			disabled = append(disabled, skill.Name)
		}
		s.Settings.Skills.Disabled = disabled
		s.SettingsChanged()
	})
	r.remove = components.NewIconButton("trash-2", "Remove skill", func() {
		if err := skills.Remove(s.SkillsDir, skill.Name); err != nil {
			s.Error("Couldn't remove the skill: " + err.Error())
		}
		p.refresh()
	}).SetSize(14)
	return r
}

func (r *skillRow) Layout(gtx giode.Context) giode.Dimensions {
	source := "your skill"
	actions := []giode.Element{r.enabled, r.remove}
	if r.skill.Source == skills.SourceProject {
		source, actions = "this project", []giode.Element{r.enabled}
	}
	return giode.Row(giode.Styles{Width: -1, Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
		Padding: giode.SymmetricInset(4, 6)},
		theme.Icon("sparkles", 15, theme.Primary),
		giode.Stack(giode.Styles{FlexGrow: 1, Gap: 2},
			giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter},
				giode.Text(r.skill.Name, giode.Styles{Color: theme.Text, FontWeight: giode.FontWeightMedium}),
				giode.Text(source, giode.Styles{Color: theme.Faint, FontSize: theme.Sizes.Tiny})),
			giode.Text(r.skill.Description, giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Small, MaxLines: 2,
				TextOverflow: giode.TextOverflowEllipsis}),
		),
		giode.Row(giode.Styles{Gap: 4, Align: giode.AlignCenter}, actions...),
	).Layout(gtx)
}
