package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"gocode/internal/llm"
)

// WriteFile creates or overwrites a file.
type WriteFile struct{}

type writeFileArgs struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// Spec describes write_file to the model.
func (WriteFile) Spec() llm.ToolSpec {
	return llm.ToolSpec{
		Name: "write_file",
		Description: "Write a file, replacing its whole content, or create it (and any missing parent directories). " +
			"Prefer edit_file for changing part of an existing file; read an existing file before overwriting it. " +
			"An existing file keeps its permissions.",
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "path": {"type": "string", "description": "File path, absolute or relative to the project root."},
    "content": {"type": "string", "description": "The complete new content of the file."}
  },
  "required": ["path", "content"],
  "additionalProperties": false
}`),
	}
}

// Access reports that write_file changes files.
func (WriteFile) Access() Access { return AccessWrite }

// Summary names the file being written.
func (WriteFile) Summary(raw json.RawMessage) string {
	var args writeFileArgs
	summaryArgs(raw, &args)
	return "Write " + args.Path
}

// Preview shows the change as a unified diff.
func (WriteFile) Preview(env Env, raw json.RawMessage) (string, error) {
	abs, args, before, err := prepareWrite(env, raw)
	if err != nil {
		return "", err
	}
	return fileDiff(env, abs, before, args.Content), nil
}

// Run writes the file.
func (WriteFile) Run(ctx context.Context, env Env, raw json.RawMessage) Result {
	abs, args, before, err := prepareWrite(env, raw)
	if err != nil {
		return Errorf("%v", err)
	}
	if err := writeText(abs, args.Content, before.mode); err != nil {
		return Errorf("cannot write %s: %v", args.Path, err)
	}
	verb := "Created"
	if before.exists {
		verb = "Overwrote"
	}
	return Result{
		Output:  fmt.Sprintf("%s %s (%d bytes)", verb, displayPath(env, abs), len(args.Content)),
		Display: fileDiff(env, abs, before, args.Content),
	}
}

// prepareWrite decodes the arguments and loads the file being replaced.
func prepareWrite(env Env, raw json.RawMessage) (string, writeFileArgs, existingFile, error) {
	var args writeFileArgs
	if err := decodeArgs(raw, &args); err != nil {
		return "", args, existingFile{}, err
	}
	abs, err := resolvePath(env, args.Path)
	if err != nil {
		return "", args, existingFile{}, err
	}
	before, err := loadExisting(abs)
	return abs, args, before, err
}
