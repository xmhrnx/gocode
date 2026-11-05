package skills

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// urlSearchDepth reaches skills/<name>/SKILL.md in a repository and the
// same inside the top folder of a GitHub source archive.
const urlSearchDepth = 3

// InstallFromURL installs the skills found at rawURL: a .zip download, or a
// git repository cloned with git clone --depth 1 (a GitHub URL may point at
// a folder with /tree/<branch>/<folder>). Every skill folder found is
// installed; the ones that succeeded are returned even when others failed.
func InstallFromURL(ctx context.Context, client *http.Client, rawURL, userDir string, overwrite bool) ([]Skill, error) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return nil, err
	}
	// Only web URLs: git also accepts file:// and ext:: transports, which
	// could read local files or run commands.
	if u.Scheme != "https" && u.Scheme != "http" {
		return nil, fmt.Errorf("unsupported URL %q; use an http or https link", rawURL)
	}
	tmp, err := os.MkdirTemp("", "gocode-skill-url-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp) // best effort: only a temporary copy remains
	root, err := fetchSource(ctx, client, u, tmp)
	if err != nil {
		return nil, err
	}
	dirs, err := findSkillDirs(root, urlSearchDepth)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", rawURL, err)
	}
	return installAll(dirs, userDir, overwrite)
}

// fetchSource downloads and unpacks u into tmp and returns the folder to
// search for skills.
func fetchSource(ctx context.Context, client *http.Client, u *url.URL, tmp string) (string, error) {
	if !strings.HasSuffix(strings.ToLower(u.Path), ".zip") {
		source, err := parseRepoURL(u)
		if err != nil {
			return "", err
		}
		return cloneRepo(ctx, source, tmp)
	}
	archive := filepath.Join(tmp, "download.zip")
	if err := downloadFile(ctx, client, u.String(), archive); err != nil {
		return "", err
	}
	root := filepath.Join(tmp, "unpacked")
	if err := extractZip(archive, root); err != nil {
		return "", err
	}
	return root, nil
}

// installAll installs each skill folder, collecting failures so one bad
// skill does not block the others.
func installAll(dirs []string, userDir string, overwrite bool) ([]Skill, error) {
	var installed []Skill
	var errs []error
	for _, dir := range dirs {
		skill, err := InstallFromFolder(dir, userDir, overwrite)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		installed = append(installed, skill)
	}
	return installed, errors.Join(errs...)
}
