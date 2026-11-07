package settingsview

import "gocode/ui/core"

func generalEntries(s *core.Services) []entry {
	g := &s.Settings.General
	return []entry{
		{title: "Default project folder", description: "Where new chats start. Empty uses the folder GoCode was started in.",
			keywords: "directory working dir path cwd", control: text(s, &g.DefaultWorkDir, "e.g. ~/code/project", 300)},
		{title: "Reopen tabs from last time", description: "Start with the chats and settings that were open when GoCode closed.",
			keywords: "restore tabs session startup reopen", control: switchControl(s, &g.RestoreTabs)},
		{title: "Name chats automatically", description: "Ask the model for a short title after the first message.",
			keywords: "title rename auto", control: switchControl(s, &g.AutoTitle)},
		{title: "Confirm quitting while working", description: "Ask before closing the window while an agent or command is running.",
			keywords: "close exit quit busy", control: switchControl(s, &g.ConfirmCloseBusy)},
	}
}
