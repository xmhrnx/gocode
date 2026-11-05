package tools

import (
	"context"
	"encoding/json"
	"os"

	"gocode/internal/llm"
)

// ReadFile reads a text file with line numbers.
type ReadFile struct{}

type readFileArgs struct {
	Path   string `json:"path"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}

const (
	defaultReadLimit = 2000
	maxReadLineBytes = 2000
)

// Spec describes read_file to the model.
func (ReadFile) Spec() llm.ToolSpec {
	return llm.ToolSpec{
		Name: "read_file",
		Description: "Read a text file. Returns lines prefixed with their line number (right-aligned) and a tab, like `cat -n`; " +
			"the prefix is not part of the file. Reads up to 2000 lines from `offset`; page through longer files with `offset` and `limit`. " +
			"Lines over 2000 characters are truncated. Binary files and directories are refused (use list_dir for directories). " +
			"Read a file before editing it.",
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "path": {"type": "string", "description": "File path, absolute or relative to the project root."},
    "offset": {"type": "integer", "minimum": 1, "description": "1-based line number to start reading at. Defaults to 1."},
    "limit": {"type": "integer", "minimum": 1, "description": "Maximum number of lines to return. Defaults to 2000."}
  },
  "required": ["path"],
  "additionalProperties": false
}`),
	}
}

// Access reports that read_file only reads.
func (ReadFile) Access() Access { return AccessRead }

// Summary names the file being read.
func (ReadFile) Summary(raw json.RawMessage) string {
	var args readFileArgs
	summaryArgs(raw, &args)
	return "Read " + args.Path
}

// Run reads the requested lines.
func (ReadFile) Run(ctx context.Context, env Env, raw json.RawMessage) Result {
	var args readFileArgs
	if err := decodeArgs(raw, &args); err != nil {
		return Errorf("%v", err)
	}
	if args.Offset < 0 || args.Limit < 0 {
		return Errorf("offset and limit must be positive")
	}
	abs, err := resolvePath(env, args.Path)
	if err != nil {
		return Errorf("%v", err)
	}
	if msg := checkReadable(abs); msg != "" {
		return Errorf("%s", msg)
	}
	f, err := os.Open(abs)
	if err != nil {
		return Errorf("cannot open %s: %v", args.Path, err)
	}
	defer f.Close()
	offset, limit := max(args.Offset, 1), args.Limit
	if limit == 0 {
		limit = defaultReadLimit
	}
	out, err := numberedLines(f, offset, limit)
	if err != nil {
		return Errorf("reading %s: %v", args.Path, err)
	}
	return Result{Output: truncateHead(out, outputLimit(env))}
}
