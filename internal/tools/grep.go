package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"

	"gocode/internal/llm"
)

// Grep searches file contents with a regular expression.
type Grep struct{}

type grepArgs struct {
	Pattern         string `json:"pattern"`
	Path            string `json:"path"`
	Include         string `json:"include"`
	CaseInsensitive bool   `json:"case_insensitive"`
	OutputMode      string `json:"output_mode"`
	Context         int    `json:"context"`
}

// Spec describes grep to the model.
func (Grep) Spec() llm.ToolSpec {
	return llm.ToolSpec{
		Name: "grep",
		Description: "Search file contents with a regular expression (Go RE2 syntax: no lookaround or backreferences; escape literal braces and dots). " +
			"Each line is matched separately. In \"content\" mode prints matching lines as `path:line: text` " +
			"(context lines as `path-line- text`, groups separated by `--`); in \"files\" mode prints only the paths of matching files. " +
			"Stops after 300 matches. Skips binary files, .git, node_modules, virtualenvs, caches and paths ignored by the root .gitignore.",
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "pattern": {"type": "string", "description": "RE2 regular expression to search for, e.g. \"func\\\\s+New\\\\w*\"."},
    "path": {"type": "string", "description": "File or directory to search, absolute or relative to the project root. Defaults to the project root."},
    "include": {"type": "string", "description": "Only search files matching this glob, e.g. \"*.go\" or \"*.{ts,tsx}\". Without a \"/\" it matches file names at any depth; with one it matches paths relative to path."},
    "case_insensitive": {"type": "boolean", "description": "Ignore letter case. Defaults to false."},
    "output_mode": {"type": "string", "enum": ["content", "files"], "description": "\"content\" (default) shows matching lines; \"files\" lists matching file paths only."},
    "context": {"type": "integer", "minimum": 0, "maximum": 5, "description": "Lines of context to show before and after each match in content mode. Defaults to 0."}
  },
  "required": ["pattern"],
  "additionalProperties": false
}`),
	}
}

// Access reports that grep only reads.
func (Grep) Access() Access { return AccessRead }

// Summary shows the pattern and where it is searched.
func (Grep) Summary(raw json.RawMessage) string {
	var args grepArgs
	summaryArgs(raw, &args)
	if args.Path != "" {
		return fmt.Sprintf("Grep %s in %s", shorten(args.Pattern, 60), args.Path)
	}
	return "Grep " + shorten(args.Pattern, 60)
}

// Run performs the search.
func (Grep) Run(ctx context.Context, env Env, raw json.RawMessage) Result {
	var args grepArgs
	if err := decodeArgs(raw, &args); err != nil {
		return Errorf("%v", err)
	}
	search, err := newGrepSearch(args)
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
	err = walkFiles(ctx, base, loadIgnoreRules(env.WorkDir), func(path string, d fs.DirEntry) error {
		rel, err := filepath.Rel(base, path)
		if rel == "." {
			// path is the file the model asked to search.
			rel = filepath.Base(path)
		}
		if err == nil && search.wants(filepath.ToSlash(rel)) {
			search.searchFile(path, displayPath(env, path))
		}
		if search.capped {
			return fs.SkipAll
		}
		return nil
	})
	if err != nil {
		return Errorf("cannot search %s: %v", args.Path, err)
	}
	return Result{Output: truncateHead(search.result(), outputLimit(env))}
}

// newGrepSearch validates the arguments and compiles the patterns.
func newGrepSearch(args grepArgs) (*grepSearch, error) {
	if args.Pattern == "" {
		return nil, fmt.Errorf("pattern is required")
	}
	if args.OutputMode != "" && args.OutputMode != "content" && args.OutputMode != "files" {
		return nil, fmt.Errorf("output_mode must be \"content\" or \"files\", not %q", args.OutputMode)
	}
	if args.Context < 0 || args.Context > 5 {
		return nil, fmt.Errorf("context must be between 0 and 5, not %d", args.Context)
	}
	expr := args.Pattern
	if args.CaseInsensitive {
		expr = "(?i)" + expr
	}
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, fmt.Errorf("invalid regular expression (Go RE2 syntax): %v", err)
	}
	search := &grepSearch{re: re, context: args.Context, filesOnly: args.OutputMode == "files"}
	if args.Include != "" {
		include := filepath.ToSlash(args.Include)
		if search.include, err = compileGlob(include); err != nil {
			return nil, err
		}
		search.includeBase = !strings.Contains(include, "/")
	}
	return search, nil
}
