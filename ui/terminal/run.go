package terminal

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gocode/internal/shell"
)

// submit runs the typed command as a new block.
func (t *Tab) submit() {
	command := strings.TrimSpace(t.input.Text())
	if command == "" || t.running() != nil {
		return
	}
	t.input.SetText("")
	t.matches = nil
	t.history.add(command)
	if command == "clear" || command == "cls" {
		t.clear()
		return
	}
	t.run(command)
}

// run starts command in a new block.
func (t *Tab) run(command string) {
	ctx, cancel := context.WithCancel(context.Background())
	t.nextID++
	b := &block{id: t.nextID, command: command, workDir: t.session.WorkDir(),
		started: time.Now(), running: true, cancel: cancel}
	t.blocks = append(t.blocks, b)
	app, maxLines := t.services.App, t.services.Settings.Terminal.MaxBlockLines
	go func() {
		result, err := t.session.Run(ctx, command, func(chunk string) {
			app.Post(func() { b.appendOutput(chunk, maxLines) })
		})
		app.Post(func() {
			b.running, b.exitCode, b.duration = false, result.ExitCode, time.Since(b.started)
			if err != nil && ctx.Err() == nil {
				b.appendOutput("\n"+err.Error()+"\n", maxLines)
			}
			b.revision++
		})
	}()
	t.list.ScrollToBottom()
}

// stop cancels the running command.
func (t *Tab) stop() {
	if b := t.running(); b != nil {
		b.cancel()
	}
}

// clear removes finished blocks.
func (t *Tab) clear() {
	if t.running() == nil {
		t.blocks = nil
	}
}

// recall steps through earlier commands with Up and Down.
func (t *Tab) recall(up bool) {
	if command, ok := t.history.step(up); ok {
		t.setLine(command)
	}
}

// complete fills in the last word of the command, as far as the
// matching files or commands agree, and lists them when there are
// several.
func (t *Tab) complete() {
	line := t.input.Text()
	result := shell.Complete(line, t.session.WorkDir())
	if result.Line != line {
		t.setLine(result.Line)
	}
	t.matches = result.Matches
}

// setLine replaces the command with line, leaving the caret at its end.
func (t *Tab) setLine(line string) {
	t.input.SelectAll()
	t.input.InsertText(line)
}

func startDir(configured string) string {
	if configured != "" {
		return configured
	}
	if dir, err := os.Getwd(); err == nil && dir != "/" {
		return dir
	}
	home, _ := os.UserHomeDir()
	return home
}

func baseName(dir string) string {
	if name := filepath.Base(dir); name != "." && name != string(filepath.Separator) {
		return name
	}
	return dir
}
