package tools

import (
	"bufio"
	"os"
	"path"
	"strings"
)

// gitignorePattern is one line of a .gitignore, supporting the common
// subset: plain names, globs like "*.log", "dir/" for directories only and
// a leading "/" (or an inner "/") to anchor at the root. Negations are not
// supported and are skipped by the parser.
type gitignorePattern struct {
	glob     string
	dirOnly  bool
	anchored bool
}

// parseGitignore reads patterns from the .gitignore file at file. A missing
// file yields no patterns.
func parseGitignore(file string) ([]gitignorePattern, error) {
	f, err := os.Open(file)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var patterns []gitignorePattern
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if p, ok := parseGitignoreLine(scanner.Text()); ok {
			patterns = append(patterns, p)
		}
	}
	return patterns, scanner.Err()
}

// parseGitignoreLine parses one line, reporting false for blank lines,
// comments and the unsupported negations.
func parseGitignoreLine(line string) (gitignorePattern, bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
		return gitignorePattern{}, false
	}
	var p gitignorePattern
	if strings.HasSuffix(line, "/") {
		p.dirOnly = true
		line = strings.TrimRight(line, "/")
	}
	// "**/name" matches name at any depth, exactly like an unanchored name.
	line = strings.TrimPrefix(line, "**/")
	if strings.HasPrefix(line, "/") || strings.Contains(line, "/") {
		p.anchored = true
		line = strings.TrimLeft(line, "/")
	}
	if line == "" {
		return gitignorePattern{}, false
	}
	p.glob = line
	return p, true
}

// matches reports whether the slash-separated path rel (relative to the
// .gitignore's directory) is ignored by p.
func (p gitignorePattern) matches(rel string, isDir bool) bool {
	if p.dirOnly && !isDir {
		return false
	}
	if p.anchored {
		ok, _ := path.Match(p.glob, rel)
		return ok
	}
	ok, _ := path.Match(p.glob, path.Base(rel))
	return ok
}
