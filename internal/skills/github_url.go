package skills

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

// repoSource is a git repository to clone and the folder of interest in it.
type repoSource struct {
	// CloneURL is passed to git clone.
	CloneURL string
	// Branch is the branch or tag to clone; empty means the default branch.
	Branch string
	// Subdir is the slash-separated folder inside the repository to search
	// for skills; empty means the whole repository.
	Subdir string
}

// parseRepoURL turns a repository URL into what to clone. GitHub web URLs
// may point into the repository with /tree/<branch>/<subdir>; other hosts
// are cloned as given. A branch name containing "/" cannot be told apart
// from the subdir and is not supported.
func parseRepoURL(u *url.URL) (repoSource, error) {
	host := strings.TrimPrefix(strings.ToLower(u.Host), "www.")
	if host != "github.com" {
		return repoSource{CloneURL: u.String()}, nil
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return repoSource{}, fmt.Errorf("%s is not a GitHub repository URL (https://github.com/owner/repo)", u)
	}
	source := repoSource{
		CloneURL: "https://github.com/" + parts[0] + "/" + strings.TrimSuffix(parts[1], ".git") + ".git",
	}
	rest := parts[2:]
	if len(rest) == 0 {
		return source, nil
	}
	if rest[0] != "tree" || len(rest) < 2 || rest[1] == "" {
		return repoSource{}, fmt.Errorf("unsupported GitHub URL %s; use https://github.com/owner/repo or .../tree/<branch>/<folder>", u)
	}
	source.Branch = rest[1]
	source.Subdir = strings.Join(rest[2:], "/")
	if source.Subdir != "" && !filepath.IsLocal(filepath.FromSlash(source.Subdir)) {
		return repoSource{}, fmt.Errorf("invalid folder %q in %s", source.Subdir, u)
	}
	return source, nil
}
