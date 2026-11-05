package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gocode/internal/llm"
)

// ListDir shows a directory tree.
type ListDir struct{}

type listDirArgs struct {
	Path  string `json:"path"`
	Depth int    `json:"depth"`
}

const (
	defaultListDepth = 2
	maxListDepth     = 5
	maxListEntries   = 500
)

// Spec describes list_dir to the model.
func (ListDir) Spec() llm.ToolSpec {
	return llm.ToolSpec{
		Name: "list_dir",
		Description: "List a directory as an indented tree. Directories end with \"/\" and come before files; entries are sorted by name. " +
			"Skips .git, node_modules, virtualenvs, caches and paths ignored by the root .gitignore. Stops after 500 entries. " +
			"Use glob to find files by name pattern.",
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "path": {"type": "string", "description": "Directory path, absolute or relative to the project root. Defaults to the project root."},
    "depth": {"type": "integer", "minimum": 1, "maximum": 5, "description": "How many levels deep to list; 1 lists only direct children. Defaults to 2."}
  },
  "additionalProperties": false
}`),
	}
}

// Access reports that list_dir only reads.
func (ListDir) Access() Access { return AccessRead }

// Summary names the directory being listed.
func (ListDir) Summary(raw json.RawMessage) string {
	var args listDirArgs
	summaryArgs(raw, &args)
	if args.Path == "" {
		args.Path = "."
	}
	return "List " + args.Path
}

// Run lists the tree.
func (ListDir) Run(ctx context.Context, env Env, raw json.RawMessage) Result {
	var args listDirArgs
	if err := decodeArgs(raw, &args); err != nil {
		return Errorf("%v", err)
	}
	if args.Path == "" {
		args.Path = "."
	}
	depth := args.Depth
	if depth <= 0 {
		depth = defaultListDepth
	}
	depth = min(depth, maxListDepth)
	abs, err := resolvePath(env, args.Path)
	if err != nil {
		return Errorf("%v", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return Errorf("cannot list %s: %v", args.Path, err)
	}
	if !info.IsDir() {
		return Errorf("%s is a file, not a directory; use read_file to read it", args.Path)
	}
	tree := &treeLister{ctx: ctx, rules: loadIgnoreRules(env.WorkDir), maxDepth: depth}
	tree.lines = append(tree.lines, strings.TrimSuffix(displayPath(env, abs), "/")+"/")
	tree.list(abs, 1)
	if err := ctx.Err(); err != nil {
		return Errorf("listing cancelled: %v", err)
	}
	if tree.capped {
		tree.lines = append(tree.lines, fmt.Sprintf("[listing stopped at %d entries; list a subdirectory or lower depth]", maxListEntries))
	}
	return Result{Output: truncateHead(strings.Join(tree.lines, "\n"), outputLimit(env))}
}
