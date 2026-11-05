package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gocode/internal/llm"
)

// Glob finds files by name pattern.
type Glob struct{}

type globArgs struct {
	Pattern string `json:"pattern"`
	Path    string `json:"path"`
}

// globMatch is a matching file and when it last changed.
type globMatch struct {
	path    string
	modTime time.Time
}

const maxGlobResults = 200

// Spec describes glob to the model.
func (Glob) Spec() llm.ToolSpec {
	return llm.ToolSpec{
		Name: "glob",
		Description: "Find files whose path matches a glob pattern, e.g. \"**/*.go\" or \"src/**/test_*.py\". " +
			"The pattern is matched against paths relative to `path`: \"*\" does not cross \"/\", so \"*.go\" only matches the top level " +
			"and \"**/\" matches any number of directories; \"{a,b}\" and \"[abc]\" are supported. " +
			"Returns up to 200 file paths (relative to the project root), most recently modified first. " +
			"Skips .git, node_modules, virtualenvs, caches and paths ignored by the root .gitignore.",
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "pattern": {"type": "string", "description": "Glob pattern relative to path, e.g. \"**/*.ts\"."},
    "path": {"type": "string", "description": "Directory to search in, absolute or relative to the project root. Defaults to the project root."}
  },
  "required": ["pattern"],
  "additionalProperties": false
}`),
	}
}

// Access reports that glob only reads.
func (Glob) Access() Access { return AccessRead }

// Summary shows the pattern and where it is searched.
func (Glob) Summary(raw json.RawMessage) string {
	var args globArgs
	summaryArgs(raw, &args)
	if args.Path != "" {
		return fmt.Sprintf("Glob %s in %s", args.Pattern, args.Path)
	}
	return "Glob " + args.Pattern
}

// Run walks the tree and lists matching files.
func (Glob) Run(ctx context.Context, env Env, raw json.RawMessage) Result {
	var args globArgs
	if err := decodeArgs(raw, &args); err != nil {
		return Errorf("%v", err)
	}
	if args.Pattern == "" || filepath.IsAbs(args.Pattern) {
		return Errorf("pattern must be a non-empty relative glob; put the directory to search in path")
	}
	re, err := compileGlob(strings.TrimPrefix(filepath.ToSlash(args.Pattern), "./"))
	if err != nil {
		return Errorf("%v", err)
	}
	if args.Path == "" {
		args.Path = "."
	}
	base, err := resolvePath(env, args.Path)
	if err != nil {
		return Errorf("%v", err)
	}
	matches, err := findGlobMatches(ctx, env, base, re)
	if err != nil {
		return Errorf("cannot search %s: %v", args.Path, err)
	}
	return Result{Output: truncateHead(formatGlobMatches(matches), outputLimit(env))}
}

// findGlobMatches walks base for files whose relative path matches re.
func findGlobMatches(ctx context.Context, env Env, base string, re *regexp.Regexp) ([]globMatch, error) {
	var matches []globMatch
	err := walkFiles(ctx, base, loadIgnoreRules(env.WorkDir), func(path string, d fs.DirEntry) error {
		rel, err := filepath.Rel(base, path)
		if err != nil || !re.MatchString(filepath.ToSlash(rel)) {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		matches = append(matches, globMatch{path: displayPath(env, path), modTime: info.ModTime()})
		return nil
	})
	return matches, err
}

// formatGlobMatches lists the newest matches first, capped.
func formatGlobMatches(matches []globMatch) string {
	if len(matches) == 0 {
		return "No files found."
	}
	sort.SliceStable(matches, func(i, j int) bool { return matches[i].modTime.After(matches[j].modTime) })
	var out strings.Builder
	for _, m := range matches[:min(len(matches), maxGlobResults)] {
		out.WriteString(m.path + "\n")
	}
	if len(matches) > maxGlobResults {
		fmt.Fprintf(&out, "[%d more files not shown; use a more specific pattern or path]\n", len(matches)-maxGlobResults)
	}
	return strings.TrimSuffix(out.String(), "\n")
}
