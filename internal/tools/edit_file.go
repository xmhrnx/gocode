package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"gocode/internal/llm"
)

// EditFile replaces exact text in a file.
type EditFile struct{}

type editFileArgs struct {
	Path       string `json:"path"`
	OldString  string `json:"old_string"`
	NewString  string `json:"new_string"`
	ReplaceAll bool   `json:"replace_all"`
}

// editPlan is a computed edit, ready to preview or write.
type editPlan struct {
	abs          string
	before       existingFile
	after        string
	replacements int
}

// Spec describes edit_file to the model.
func (EditFile) Spec() llm.ToolSpec {
	return llm.ToolSpec{
		Name: "edit_file",
		Description: "Replace exact text in a file. `old_string` must match the file exactly, including whitespace and indentation " +
			"(copy it from read_file output without the line-number prefix) and must be unique in the file unless `replace_all` is true; " +
			"include enough surrounding lines to make it unique. Line endings are handled: LF text matches a CRLF file. " +
			"An empty `old_string` creates a new file containing `new_string`. Read the file before editing it.",
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "path": {"type": "string", "description": "File path, absolute or relative to the project root."},
    "old_string": {"type": "string", "description": "The exact text to replace. Empty only to create a new file."},
    "new_string": {"type": "string", "description": "The text to put in its place; must differ from old_string."},
    "replace_all": {"type": "boolean", "description": "Replace every occurrence of old_string instead of requiring exactly one. Defaults to false."}
  },
  "required": ["path", "old_string", "new_string"],
  "additionalProperties": false
}`),
	}
}

// Access reports that edit_file changes files.
func (EditFile) Access() Access { return AccessWrite }

// Summary names the file being edited.
func (EditFile) Summary(raw json.RawMessage) string {
	var args editFileArgs
	summaryArgs(raw, &args)
	return "Edit " + args.Path
}

// Preview shows the edit as a unified diff.
func (EditFile) Preview(env Env, raw json.RawMessage) (string, error) {
	plan, err := planEdit(env, raw)
	if err != nil {
		return "", err
	}
	return fileDiff(env, plan.abs, plan.before, plan.after), nil
}

// Run applies the edit.
func (EditFile) Run(ctx context.Context, env Env, raw json.RawMessage) Result {
	plan, err := planEdit(env, raw)
	if err != nil {
		return Errorf("%v", err)
	}
	if err := writeText(plan.abs, plan.after, plan.before.mode); err != nil {
		return Errorf("cannot write %s: %v", plan.abs, err)
	}
	name := displayPath(env, plan.abs)
	output := fmt.Sprintf("Edited %s (%d replacement(s))", name, plan.replacements)
	if !plan.before.exists {
		output = "Created " + name
	}
	return Result{Output: output, Display: fileDiff(env, plan.abs, plan.before, plan.after)}
}

// planEdit validates the arguments and computes the new file content.
func planEdit(env Env, raw json.RawMessage) (editPlan, error) {
	var args editFileArgs
	if err := decodeArgs(raw, &args); err != nil {
		return editPlan{}, err
	}
	if args.OldString == args.NewString {
		return editPlan{}, errors.New("old_string and new_string are identical, so there is nothing to change")
	}
	abs, err := resolvePath(env, args.Path)
	if err != nil {
		return editPlan{}, err
	}
	before, err := loadExisting(abs)
	if err != nil {
		return editPlan{}, err
	}
	plan := editPlan{abs: abs, before: before}
	switch {
	case args.OldString == "" && before.content == "":
		plan.after, plan.replacements = args.NewString, 1
	case args.OldString == "":
		return editPlan{}, fmt.Errorf("old_string is empty but %s already has content; give the exact text to replace, or use write_file to overwrite the file", args.Path)
	case !before.exists:
		return editPlan{}, fmt.Errorf("file not found: %s; to create it, pass an empty old_string or use write_file", abs)
	default:
		plan.after, plan.replacements, err = applyEdit(before.content, args.OldString, args.NewString, args.ReplaceAll)
	}
	return plan, err
}
