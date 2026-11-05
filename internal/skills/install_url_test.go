package skills

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// TestInstallFromURLZip serves a GitHub-style source archive locally, so
// the download path is covered without reaching the network.
func TestInstallFromURLZip(t *testing.T) {
	archive := buildZip(t, map[string]string{
		"repo-main/README.md":                  "repo",
		"repo-main/skills/alpha/" + SkillFile:  skillMarkdown("alpha", "a"),
		"repo-main/skills/beta/" + SkillFile:   skillMarkdown("beta", "b"),
		"repo-main/skills/beta/scripts/run.sh": "echo",
	})
	data, err := os.ReadFile(archive)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/main.zip" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(data) // a failed write surfaces as a failed install
	}))
	defer server.Close()

	userDir := t.TempDir()
	installed, err := InstallFromURL(context.Background(), server.Client(), server.URL+"/main.zip", userDir, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(installed) != 2 || installed[0].Name != "alpha" || installed[1].Name != "beta" {
		t.Errorf("installed = %v, want alpha and beta", installed)
	}

	_, err = InstallFromURL(context.Background(), server.Client(), server.URL+"/missing.zip", userDir, false)
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Errorf("error = %v, want a 404", err)
	}
}

func TestInstallFromURLRejectsSchemes(t *testing.T) {
	for _, rawURL := range []string{"file:///etc", "ext::sh -c id", "git@github.com:o/r.git"} {
		if _, err := InstallFromURL(context.Background(), http.DefaultClient, rawURL, t.TempDir(), false); err == nil {
			t.Errorf("%q was accepted", rawURL)
		}
	}
}
