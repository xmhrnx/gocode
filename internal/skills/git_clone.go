package skills

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// cloneRepo shallow-clones source into dir/repo and returns the folder to
// search for skills.
func cloneRepo(ctx context.Context, source repoSource, dir string) (string, error) {
	dest := filepath.Join(dir, "repo")
	args := []string{"clone", "--depth", "1"}
	if source.Branch != "" {
		args = append(args, "--branch", source.Branch)
	}
	// "--" stops a URL that starts with "-" from being read as an option.
	args = append(args, "--", source.CloneURL, dest)
	cmd := exec.CommandContext(ctx, "git", args...)
	// Without a terminal, a credential prompt for a private or missing
	// repository would hang the install instead of failing.
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git clone %s: %w: %s", source.CloneURL, err, strings.TrimSpace(string(output)))
	}
	return filepath.Join(dest, filepath.FromSlash(source.Subdir)), nil
}
