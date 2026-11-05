package skills

import (
	"fmt"
	"strings"
)

// promptIntro tells the model what the listed skills are for and how to
// use one; the listing itself stays short so many skills cost few tokens.
const promptIntro = `# Skills
You can load these skills with the skill tool when a task matches their description. ` +
	`A skill holds instructions, and sometimes scripts or reference files, for a specific kind of task. ` +
	`Load a matching skill before starting such a task and follow its instructions.`

// PromptSection returns the system prompt section listing skills by name
// and description, or "" when there are none.
func PromptSection(skills []Skill) string {
	if len(skills) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(promptIntro)
	for _, skill := range skills {
		fmt.Fprintf(&b, "\n- %s: %s", skill.Name, oneLine(skill.Description))
	}
	return b.String()
}

// Enabled returns the skills whose names are not in disabled (the
// settings.Skills.Disabled list), keeping their order.
func Enabled(skills []Skill, disabled []string) []Skill {
	off := make(map[string]bool, len(disabled))
	for _, name := range disabled {
		off[name] = true
	}
	var enabled []Skill
	for _, skill := range skills {
		if !off[skill.Name] {
			enabled = append(enabled, skill)
		}
	}
	return enabled
}

// oneLine collapses a (possibly folded) description onto one line so each
// skill stays a single list item.
func oneLine(text string) string {
	return strings.Join(strings.Fields(text), " ")
}
