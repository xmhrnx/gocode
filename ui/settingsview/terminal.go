package settingsview

import "gocode/ui/core"

func terminalEntries(s *core.Services) []entry {
	t := &s.Settings.Terminal
	return []entry{
		{title: "Shell", description: "The program that runs commands, for the terminal and the agent. Empty uses $SHELL, or PowerShell on Windows.",
			keywords: "bash zsh powershell sh shell program", control: text(s, &t.Shell, "/bin/zsh", 220)},
		{title: "Lines kept per block", description: "Older output of long-running commands is dropped beyond this, to keep memory low.",
			keywords: "scrollback output lines memory", control: number(s, &t.MaxBlockLines, 100, 100000)},
	}
}
