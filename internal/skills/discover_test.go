package skills

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDiscoverPrecedence(t *testing.T) {
	userDir, workDir := t.TempDir(), t.TempDir()
	claudeDir := filepath.Join(workDir, ".claude", "skills")
	gocodeDir := filepath.Join(workDir, ".gocode", "skills")
	writeSkill(t, userDir, "only-user", "only-user", "user only")
	writeSkill(t, userDir, "shared", "shared", "from user")
	writeSkill(t, userDir, "both-projects", "both-projects", "from user")
	writeSkill(t, claudeDir, "shared", "shared", "from claude")
	writeSkill(t, claudeDir, "both-projects", "both-projects", "from claude")
	writeSkill(t, gocodeDir, "both", "both-projects", "from gocode")
	writeSkill(t, userDir, ".installing-123", "hidden", "staging folder")
	writeFiles(t, userDir, map[string]string{"notes/readme.txt": "not a skill"})

	found, problems := Discover(userDir, workDir)
	if len(problems) != 0 {
		t.Fatalf("problems = %v", problems)
	}
	tests := []struct {
		name   string
		desc   string
		source Source
		dir    string
	}{
		{"both-projects", "from gocode", SourceProject, filepath.Join(gocodeDir, "both")},
		{"only-user", "user only", SourceUser, filepath.Join(userDir, "only-user")},
		{"shared", "from claude", SourceProject, filepath.Join(claudeDir, "shared")},
	}
	if len(found) != len(tests) {
		t.Fatalf("found %d skills, want %d: %v", len(found), len(tests), found)
	}
	for i, tt := range tests {
		want := Skill{Name: tt.name, Description: tt.desc, Dir: tt.dir, Source: tt.source}
		if found[i] != want {
			t.Errorf("skill %d = %+v, want %+v", i, found[i], want)
		}
	}
}

func TestDiscoverReportsInvalidSkills(t *testing.T) {
	userDir := t.TempDir()
	writeSkill(t, userDir, "good", "good", "fine")
	writeFiles(t, userDir, map[string]string{"bad/" + SkillFile: "---\nname: Bad Name\ndescription: d\n---\n"})

	found, problems := Discover(userDir, "")
	if len(found) != 1 || found[0].Name != "good" {
		t.Errorf("found = %v, want only good", found)
	}
	if len(problems) != 1 || problems[0].Dir != filepath.Join(userDir, "bad") {
		t.Fatalf("problems = %v, want one for bad", problems)
	}
	if !strings.Contains(problems[0].Err.Error(), "lowercase") {
		t.Errorf("problem error = %v", problems[0].Err)
	}
}

func TestDiscoverMissingFolders(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "absent")
	found, problems := Discover(missing, missing)
	if len(found) != 0 || len(problems) != 0 {
		t.Errorf("got %v, %v; want nothing", found, problems)
	}
}

func TestLoad(t *testing.T) {
	dir := writeSkill(t, t.TempDir(), "tool", "tool", "d")
	writeFiles(t, dir, map[string]string{
		"scripts/run.sh":      "echo",
		"reference.md":        "ref",
		".hidden":             "x",
		".git/config":         "x",
		"templates/a/.secret": "x",
	})
	body, files, err := Load(Skill{Name: "tool", Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if body != "# tool\nDo the thing." {
		t.Errorf("body = %q", body)
	}
	want := []string{"reference.md", "scripts/run.sh"}
	if !reflect.DeepEqual(files, want) {
		t.Errorf("files = %v, want %v", files, want)
	}
}

func TestLoadCapsFileListing(t *testing.T) {
	dir := writeSkill(t, t.TempDir(), "big", "big", "d")
	extra := map[string]string{}
	for i := range maxListedFiles + 20 {
		extra[fmt.Sprintf("data/%03d.txt", i)] = "x"
	}
	writeFiles(t, dir, extra)
	_, files, err := Load(Skill{Name: "big", Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != maxListedFiles {
		t.Errorf("listed %d files, want %d", len(files), maxListedFiles)
	}
}
