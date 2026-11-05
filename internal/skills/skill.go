// Package skills discovers, installs and loads Agent Skills: folders holding
// a SKILL.md file (YAML frontmatter plus Markdown instructions) and optional
// scripts, references and templates the agent can use on demand.
package skills

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Source tells where a skill was found.
type Source string

// Skill is an installed skill.
type Skill struct {
	// Name identifies the skill; it comes from the frontmatter.
	Name string
	// Description tells the model when to use the skill.
	Description string
	// Dir is the absolute or caller-relative path of the skill folder.
	Dir string
	// Source is SourceUser or SourceProject.
	Source Source
}

// Problem is a skill folder that could not be read, kept so the UI can
// explain why a skill is missing.
type Problem struct {
	Dir string
	Err error
}

const (
	// SourceUser marks skills installed in the user's skills directory.
	SourceUser Source = "user"
	// SourceProject marks skills shipped in the project's own folders.
	SourceProject Source = "project"
)

// SkillFile is the file that makes a folder a skill.
const SkillFile = "SKILL.md"

// ErrExists is returned when installing over a skill without overwrite.
var ErrExists = errors.New("skill already installed")

// readSkill parses dir's SKILL.md and returns the skill and its body.
func readSkill(dir string, source Source) (Skill, string, error) {
	data, err := os.ReadFile(filepath.Join(dir, SkillFile))
	if err != nil {
		return Skill{}, "", err
	}
	skill, body, err := parseSkill(string(data))
	if err != nil {
		return Skill{}, "", fmt.Errorf("%s: %w", filepath.Join(dir, SkillFile), err)
	}
	skill.Dir = dir
	skill.Source = source
	return skill, body, nil
}

// parseSkill reads the frontmatter fields a skill needs and checks them;
// other keys (license, allowed-tools, metadata, ...) are accepted and ignored.
func parseSkill(content string) (Skill, string, error) {
	fields, body, err := parseFrontmatter(content)
	if err != nil {
		return Skill{}, "", err
	}
	skill := Skill{
		Name:        strings.TrimSpace(fields["name"]),
		Description: strings.TrimSpace(fields["description"]),
	}
	if err := validateName(skill.Name); err != nil {
		return Skill{}, "", err
	}
	if err := validateDescription(skill.Description); err != nil {
		return Skill{}, "", err
	}
	return skill, body, nil
}
