package skills

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallFromZip(t *testing.T) {
	tests := []struct {
		name     string
		entries  map[string]string
		wantFile string
	}{
		{
			name:     "skill at the root",
			entries:  map[string]string{SkillFile: skillMarkdown("zipped", "d"), "ref/notes.md": "notes"},
			wantFile: "ref/notes.md",
		},
		{
			name: "skill one level down",
			entries: map[string]string{
				"zipped-main/" + SkillFile:  skillMarkdown("zipped", "d"),
				"zipped-main/run.py":        "print()",
				"__MACOSX/zipped-main/._x":  "junk",
				"__MACOSX/" + SkillFile:     "not a skill",
				"zipped-main/.DS_Store":     "junk",
				"zipped-main/lib/helper.py": "pass",
			},
			wantFile: "lib/helper.py",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userDir := t.TempDir()
			skill, err := InstallFromZip(buildZip(t, tt.entries), userDir, false)
			if err != nil {
				t.Fatal(err)
			}
			if skill.Name != "zipped" || skill.Dir != filepath.Join(userDir, "zipped") {
				t.Errorf("skill = %+v", skill)
			}
			readFile(t, filepath.Join(skill.Dir, filepath.FromSlash(tt.wantFile)))
		})
	}
}

func TestInstallFromZipRejects(t *testing.T) {
	tooMany := map[string]string{SkillFile: skillMarkdown("many", "d")}
	for i := range maxInstallFiles {
		tooMany[fmt.Sprintf("f/%d", i)] = ""
	}
	tests := []struct {
		name     string
		entries  map[string]string
		wantText string
	}{
		{"zip-slip parent", map[string]string{SkillFile: skillMarkdown("x", "d"), "../evil.sh": "x"}, "outside"},
		{"zip-slip nested", map[string]string{"a/../../evil.sh": "x"}, "outside"},
		{"absolute path", map[string]string{"/etc/evil": "x"}, "outside"},
		{"no skill", map[string]string{"readme.md": "x"}, "no folder containing"},
		{"too deep", map[string]string{"a/b/" + SkillFile: skillMarkdown("x", "d")}, "no folder containing"},
		{"two skills", map[string]string{
			"one/" + SkillFile: skillMarkdown("one", "d"),
			"two/" + SkillFile: skillMarkdown("two", "d"),
		}, "contains 2 skills"},
		{"too many files", tooMany, "more than"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userDir := t.TempDir()
			_, err := InstallFromZip(buildZip(t, tt.entries), userDir, false)
			if err == nil || !strings.Contains(err.Error(), tt.wantText) {
				t.Fatalf("error = %v, want one containing %q", err, tt.wantText)
			}
			if found, _ := Discover(userDir, ""); len(found) != 0 {
				t.Errorf("a rejected archive installed %v", found)
			}
		})
	}
}

func TestBudgetRejectsLargeContent(t *testing.T) {
	limits := budget{bytes: maxInstallBytes - 3}
	err := limits.copy(&strings.Builder{}, strings.NewReader("12345"))
	if err != errTooLarge {
		t.Errorf("error = %v, want errTooLarge", err)
	}
}
