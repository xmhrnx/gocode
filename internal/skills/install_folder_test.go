package skills

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallFromFolder(t *testing.T) {
	src := writeSkill(t, t.TempDir(), "anything", "my-skill", "first")
	writeFiles(t, src, map[string]string{"scripts/run.sh": "echo hi", ".git/HEAD": "ref"})
	if err := os.Chmod(filepath.Join(src, "scripts", "run.sh"), 0o755); err != nil {
		t.Fatal(err)
	}
	userDir := filepath.Join(t.TempDir(), "skills")

	skill, err := InstallFromFolder(src, userDir, false)
	if err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(userDir, "my-skill")
	want := Skill{Name: "my-skill", Description: "first", Dir: dest, Source: SourceUser}
	if skill != want {
		t.Errorf("skill = %+v, want %+v", skill, want)
	}
	if got := readFile(t, filepath.Join(dest, "scripts", "run.sh")); got != "echo hi" {
		t.Errorf("script = %q", got)
	}
	info, err := os.Stat(filepath.Join(dest, "scripts", "run.sh"))
	if err != nil || info.Mode().Perm()&0o100 == 0 {
		t.Errorf("script lost its executable bit: %v %v", info, err)
	}
	if _, err := os.Stat(filepath.Join(dest, ".git")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf(".git was copied: %v", err)
	}
}

func TestInstallFromFolderOverwrite(t *testing.T) {
	userDir := t.TempDir()
	first := writeSkill(t, t.TempDir(), "v1", "dup", "first")
	second := writeSkill(t, t.TempDir(), "v2", "dup", "second")
	writeFiles(t, first, map[string]string{"old.txt": "stale"})
	if _, err := InstallFromFolder(first, userDir, false); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		overwrite bool
		wantErr   error
		wantDesc  string
	}{
		{"refuses without overwrite", false, ErrExists, "first"},
		{"replaces with overwrite", true, nil, "second"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := InstallFromFolder(second, userDir, tt.overwrite)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			found, _ := Discover(userDir, "")
			if len(found) != 1 || found[0].Description != tt.wantDesc {
				t.Errorf("installed = %v, want description %q", found, tt.wantDesc)
			}
		})
	}
	if _, err := os.Stat(filepath.Join(userDir, "dup", "old.txt")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("overwrite kept a file from the old version: %v", err)
	}
}

func TestInstallFromFolderRejects(t *testing.T) {
	parent := t.TempDir()
	nestedSrc := writeSkill(t, parent, "outer", "outer", "d")
	tests := []struct {
		name    string
		src     string
		userDir string
	}{
		{"no SKILL.md", t.TempDir(), t.TempDir()},
		{"source contains the skills dir", nestedSrc, filepath.Join(nestedSrc, "skills")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := InstallFromFolder(tt.src, tt.userDir, true); err == nil {
				t.Error("expected an error")
			}
		})
	}
}

func TestRemove(t *testing.T) {
	userDir := t.TempDir()
	writeSkill(t, userDir, "gone", "gone", "d")
	outside := writeSkill(t, t.TempDir(), "kept", "kept", "d")
	if err := os.Symlink(outside, filepath.Join(userDir, "linked")); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		skill   string
		wantErr bool
	}{
		{"removes a folder", "gone", false},
		{"removes only the link", "linked", false},
		{"missing skill", "absent", true},
		{"rejects traversal", "../x", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Remove(userDir, tt.skill); (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	if _, err := os.Stat(filepath.Join(outside, SkillFile)); err != nil {
		t.Errorf("removing a link deleted its target: %v", err)
	}
}
