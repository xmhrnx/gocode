package skills

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// skillRoot is a folder whose subfolders are skills.
type skillRoot struct {
	dir    string
	source Source
}

// projectSkillDirs are searched under the project root, lowest precedence
// first, so GoCode's own folder wins over the Claude Code one.
var projectSkillDirs = []string{
	filepath.Join(".claude", "skills"),
	filepath.Join(".gocode", "skills"),
}

// Discover finds the skills in userDir and in the project's .claude/skills
// and .gocode/skills folders under workDir, sorted by name. Project skills
// override user skills with the same name. Folders whose SKILL.md is invalid
// are reported as problems instead. Either directory may be empty to skip it.
func Discover(userDir, workDir string) (found []Skill, problems []Problem) {
	byName := map[string]Skill{}
	for _, root := range skillRoots(userDir, workDir) {
		skills, rootProblems := scanRoot(root)
		problems = append(problems, rootProblems...)
		for _, skill := range skills {
			byName[skill.Name] = skill // later roots take precedence
		}
	}
	for _, skill := range byName {
		found = append(found, skill)
	}
	sort.Slice(found, func(i, j int) bool { return found[i].Name < found[j].Name })
	return found, problems
}

// skillRoots lists the folders to scan, lowest precedence first.
func skillRoots(userDir, workDir string) []skillRoot {
	var roots []skillRoot
	if userDir != "" {
		roots = append(roots, skillRoot{dir: userDir, source: SourceUser})
	}
	if workDir != "" {
		for _, rel := range projectSkillDirs {
			roots = append(roots, skillRoot{dir: filepath.Join(workDir, rel), source: SourceProject})
		}
	}
	return roots
}

// scanRoot reads every skill folder directly inside root. Hidden entries
// are skipped: they include in-progress installs.
func scanRoot(root skillRoot) ([]Skill, []Problem) {
	entries, err := os.ReadDir(root.dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil // most projects have no skills folder
	}
	if err != nil {
		return nil, []Problem{{Dir: root.dir, Err: err}}
	}
	var skills []Skill
	var problems []Problem
	for _, entry := range entries {
		dir := filepath.Join(root.dir, entry.Name())
		if strings.HasPrefix(entry.Name(), ".") || !hasSkillFile(dir) {
			continue
		}
		skill, _, err := readSkill(dir, root.source)
		if err != nil {
			problems = append(problems, Problem{Dir: dir, Err: err})
			continue
		}
		skills = append(skills, skill)
	}
	return skills, problems
}

// hasSkillFile reports whether dir (or the folder it links to) holds a
// SKILL.md file.
func hasSkillFile(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, SkillFile))
	return err == nil && info.Mode().IsRegular()
}
