package settingsview

import "gocode/ui/core"

func generalEntries(s *core.Services) []entry {
	g := &s.Settings.General
	return []entry{
		{title: "Projects folder", description: "A new chat gets a folder of its own here, named after its first message, unless you open a folder for it. Empty uses ~/GoCode.",
			keywords: "directory working dir path cwd projects default folder workspace", control: text(s, &g.ProjectsDir, "~/GoCode", 300)},
		{title: "Reopen tabs from last time", description: "Start with the chats and settings that were open when GoCode closed.",
			keywords: "restore tabs session startup reopen", control: switchControl(s, &g.RestoreTabs)},
		{title: "Name chats automatically", description: "Ask the model for a short title after the first message.",
			keywords: "title rename auto", control: switchControl(s, &g.AutoTitle)},
		{title: "Confirm quitting while working", description: "Ask before closing the window while an agent or command is running.",
			keywords: "close exit quit busy", control: switchControl(s, &g.ConfirmCloseBusy)},
	}
}
