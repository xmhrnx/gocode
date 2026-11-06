package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gocode/internal/proc"
	"os/exec"
	"time"

	"gocode/internal/llm"
	"gocode/internal/shell"
)

// Bash runs a shell command.
type Bash struct{}

type bashArgs struct {
	Command        string `json:"command"`
	TimeoutSeconds int    `json:"timeout_seconds"`
	Description    string `json:"description"`
}

const (
	defaultCommandTimeout = 120
	maxCommandTimeout     = 600
	// commandWaitDelay bounds how long Run waits for output pipes after the
	// process is gone, in case a detached grandchild still holds them.
	commandWaitDelay = 2 * time.Second
)

// Spec describes bash to the model.
func (Bash) Spec() llm.ToolSpec {
	return llm.ToolSpec{
		Name: "bash",
		Description: "Run a shell command in the project root and return its combined stdout and stderr, followed by \"exit code: N\" when it fails. " +
			"Each call starts a fresh shell: `cd` and environment changes do not carry over, so chain steps with && instead. " +
			"Stdin is closed, so interactive commands fail; pass non-interactive flags. Long output is truncated in the middle, keeping the start and the end. " +
			"Commands are killed, with any processes they started, after the timeout. " +
			"Prefer read_file, list_dir, glob and grep over cat, ls, find and grep for reading and searching. " +
			"Never create or change files with the shell (heredocs, echo or printf redirection, sed -i): use write_file and edit_file, which write the exact bytes.",
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "command": {"type": "string", "description": "The command line to run."},
    "timeout_seconds": {"type": "integer", "minimum": 1, "maximum": 600, "description": "Kill the command after this many seconds. Defaults to 120."},
    "description": {"type": "string", "description": "What the command does in 5-10 words, shown to the user, e.g. \"Run the unit tests\"."}
  },
  "required": ["command"],
  "additionalProperties": false
}`),
	}
}

// Access reports that bash executes programs.
func (Bash) Access() Access { return AccessExecute }

// Summary prefers the model's description, falling back to the command.
func (Bash) Summary(raw json.RawMessage) string {
	var args bashArgs
	summaryArgs(raw, &args)
	if args.Description != "" {
		return shorten(args.Description, 80)
	}
	return "Run " + shorten(args.Command, 80)
}

// Run executes the command.
func (Bash) Run(ctx context.Context, env Env, raw json.RawMessage) Result {
	var args bashArgs
	if err := decodeArgs(raw, &args); err != nil {
		return Errorf("%v", err)
	}
	if args.Command == "" {
		return Errorf("command is required")
	}
	timeout := commandTimeout(env, args.TimeoutSeconds)
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	program := env.Shell
	if program == "" {
		program = shell.Default()
	}
	output := newHeadTailBuffer(outputLimit(env))
	scriptArgs, cleanup, err := shell.ScriptArgs(program, args.Command)
	if err != nil {
		return Errorf("cannot prepare the command: %v", err)
	}
	defer cleanup()
	cmd := exec.CommandContext(runCtx, program, scriptArgs...)
	cmd.Dir = env.WorkDir
	cmd.Stdout, cmd.Stderr = output, output
	cmd.WaitDelay = commandWaitDelay
	proc.StartGroup(cmd)
	cmd.Cancel = func() error { return proc.KillGroup(cmd) }
	err = cmd.Run()
	return commandResult(runCtx, ctx, output.String(), err, timeout)
}

// commandTimeout picks the timeout: the call's, else the environment's,
// else the default, never above the maximum.
func commandTimeout(env Env, requested int) time.Duration {
	seconds := requested
	if seconds <= 0 {
		seconds = env.CommandTimeoutSeconds
	}
	if seconds <= 0 {
		seconds = defaultCommandTimeout
	}
	return time.Duration(min(seconds, maxCommandTimeout)) * time.Second
}

// commandResult turns the command's output and outcome into a Result.
func commandResult(runCtx, parent context.Context, output string, err error, timeout time.Duration) Result {
	if output == "" {
		output = "(no output)"
	}
	var exitErr *exec.ExitError
	switch {
	case err == nil:
		return Result{Output: output}
	case errors.Is(err, exec.ErrWaitDelay):
		// The shell succeeded but left a background process holding the
		// output open; report success rather than waiting for it.
		return Result{Output: output + "\n[background processes left running; their later output is not captured]"}
	case parent.Err() != nil:
		return Result{Output: output + "\ncommand cancelled", IsError: true}
	case errors.Is(runCtx.Err(), context.DeadlineExceeded):
		return Result{Output: output + fmt.Sprintf("\ncommand timed out after %s and was killed", timeout), IsError: true}
	case errors.As(err, &exitErr):
		return Result{Output: output + fmt.Sprintf("\nexit code: %d", exitErr.ExitCode()), IsError: true}
	}
	return Errorf("cannot run command: %v", err)
}
