package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gocode/internal/llm"
	"gocode/internal/tools"
)

// skillTool lets the model load a skill's instructions on demand.
type skillTool struct {
	list func() []Skill
}

// skillArgs are the skill tool's arguments.
type skillArgs struct {
	Name string `json:"name"`
}

const toolDescription = "Load a skill: instructions, and sometimes scripts or reference files, for a specific kind of task. " +
	"Call it with a skill's name from the Skills section of the system prompt as soon as a task matches the skill's description, " +
	"then follow the returned instructions. Read the skill's files with read_file and run its scripts with bash, " +
	"using paths relative to the skill's base directory."

// NewTool returns the "skill" tool. list is called on every run, so skills
// installed or switched off since startup are picked up; it should return
// only the enabled skills.
func NewTool(list func() []Skill) tools.Tool {
	return skillTool{list: list}
}

// Spec describes the skill tool to the model.
func (skillTool) Spec() llm.ToolSpec {
	return llm.ToolSpec{
		Name:        "skill",
		Description: toolDescription,
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "name": {"type": "string", "description": "The skill's name, as listed in the system prompt."}
  },
  "required": ["name"],
  "additionalProperties": false
}`),
	}
}

// Access reports that loading a skill only reads files.
func (skillTool) Access() tools.Access { return tools.AccessRead }

// Summary names the skill being loaded.
func (skillTool) Summary(raw json.RawMessage) string {
	var args skillArgs
	_ = json.Unmarshal(raw, &args) // a summary must never fail; a bad call shows an empty name
	return "Load skill " + args.Name
}

// Run returns the skill's instructions with its base directory and files.
func (t skillTool) Run(ctx context.Context, env tools.Env, raw json.RawMessage) tools.Result {
	var args skillArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return tools.Errorf("invalid arguments (%v); pass {\"name\": \"<skill name>\"}", err)
	}
	available := t.list()
	skill, ok := findByName(available, args.Name)
	if !ok {
		return tools.Errorf("unknown skill %q; available skills: %s", args.Name, skillNames(available))
	}
	body, files, err := Load(skill)
	if err != nil {
		return tools.Errorf("load skill %s: %v", skill.Name, err)
	}
	return tools.Result{Output: renderSkill(skill, body, files)}
}

// renderSkill puts the base directory first so the model resolves the
// relative paths in the instructions against it.
func renderSkill(skill Skill, body string, files []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Base directory for this skill: %s\n\n%s", skill.Dir, body)
	if len(files) > 0 {
		b.WriteString("\n\nFiles in this skill (relative to the base directory):")
		for _, file := range files {
			b.WriteString("\n- " + file)
		}
	}
	return b.String()
}

func findByName(skills []Skill, name string) (Skill, bool) {
	for _, skill := range skills {
		if skill.Name == name {
			return skill, true
		}
	}
	return Skill{}, false
}

func skillNames(skills []Skill) string {
	if len(skills) == 0 {
		return "(none installed)"
	}
	names := make([]string, len(skills))
	for i, skill := range skills {
		names[i] = skill.Name
	}
	return strings.Join(names, ", ")
}
