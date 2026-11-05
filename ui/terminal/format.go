package terminal

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func formatDuration(d time.Duration) string {
	switch {
	case d < time.Second:
		return fmt.Sprintf("%dms", d.Milliseconds())
	case d < time.Minute:
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return d.Round(time.Second).String()
}

func shortDir(dir string) string {
	if home, err := os.UserHomeDir(); err == nil && strings.HasPrefix(dir, home) {
		return "~" + strings.TrimPrefix(dir, home)
	}
	return dir
}

const maxPromptOutput = 12000

// agentPrompt asks the agent about a block, quoting its output.
func agentPrompt(b *block) string {
	output := b.text()
	if output == "" {
		output = "(no output)"
	}
	if len(output) > maxPromptOutput {
		output = "…" + output[len(output)-maxPromptOutput:]
	}
	verdict := "It succeeded."
	if b.exitCode != 0 {
		verdict = fmt.Sprintf("It failed with exit code %d. Find the cause and fix it.", b.exitCode)
	}
	return fmt.Sprintf("I ran `%s` in %s. %s\n\nOutput:\n```\n%s\n```\n", b.command, b.workDir, verdict, output)
}
