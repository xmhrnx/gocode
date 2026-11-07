package projects

import (
	"os"
	"path/filepath"
	"testing"
)

func TestName(t *testing.T) {
	cases := map[string]string{
		"Build a snake game in Go, please": "build-a-snake-game-in",
		"  Fix the login bug!  ":           "fix-the-login-bug",
		"Écris un résumé":                  "écris-un-résumé",
		"???":                              "project",
		"":                                 "project",
	}
	for message, want := range cases {
		if got := Name(message); got != want {
			t.Errorf("Name(%q) = %q, want %q", message, got, want)
		}
	}
}

func TestCreateMakesUniqueFolders(t *testing.T) {
	parent := filepath.Join(t.TempDir(), "GoCode") // not there yet
	first, err := Create(parent, "Snake game")
	if err != nil {
		t.Fatal(err)
	}
	second, err := Create(parent, "Snake game")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(first) != "snake-game" || filepath.Base(second) != "snake-game-2" {
		t.Fatalf("made %s and %s", first, second)
	}
	if info, err := os.Stat(second); err != nil || !info.IsDir() {
		t.Fatalf("%s isn't a folder: %v", second, err)
	}
}

func TestDir(t *testing.T) {
	home, _ := os.UserHomeDir()
	if got := Dir(""); got != filepath.Join(home, "GoCode") {
		t.Errorf("default %s", got)
	}
	if got := Dir("~/code"); got != filepath.Join(home, "code") {
		t.Errorf("~ expansion %s", got)
	}
	if got := Dir("/srv/work"); got != "/srv/work" {
		t.Errorf("absolute %s", got)
	}
}

func TestDisplay(t *testing.T) {
	home, _ := os.UserHomeDir()
	if got := Display(filepath.Join(home, "GoCode")); got != filepath.Join("~", "GoCode") {
		t.Errorf("home path shows as %s", got)
	}
	if got := Display("/srv/work"); got != "/srv/work" {
		t.Errorf("other path shows as %s", got)
	}
}
