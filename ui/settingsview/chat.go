package settingsview

import (
	"gocode/internal/settings"
	"gocode/ui/core"
)

func chatEntries(s *core.Services) []entry {
	c := &s.Settings.Chat
	sendKeys := []option[settings.SendKey]{
		{settings.SendEnter, "Enter"},
		{settings.SendShiftEnter, "Shift + Enter"},
		{settings.SendCtrlEnter, "Ctrl/⌘ + Enter"},
	}
	return []entry{
		{title: "Send message with", description: "The other Enter combinations insert a new line.",
			keywords: "send key enter shift ctrl newline submit keyboard", control: choice(s, &c.SendKey, sendKeys)},
		{title: "Show reasoning", description: "Show the model's thinking, collapsed, above its answer.",
			keywords: "thinking reasoning thoughts", control: switchControl(s, &c.ShowReasoning)},
		{title: "Expand tool output", description: "Show each tool call's output or diff without clicking it.",
			keywords: "tool output diff expand", control: switchControl(s, &c.ExpandToolOutput)},
		{title: "Show token usage", description: "Show tokens sent and received, and the share served from cache.",
			keywords: "tokens usage cost cache", control: switchControl(s, &c.ShowTokenUsage)},
		{title: "Largest attachment (MB)", description: "Files and images above this size are refused.",
			keywords: "attachment upload file image size limit", control: number(s, &c.MaxAttachmentMB, 1, 200)},
		{title: "Largest inlined text file (KB)", description: "Text files up to this size are sent in full; larger ones should be read by the agent.",
			keywords: "attachment text file inline size limit", control: number(s, &c.InlineTextFilesKB, 1, 4096)},
	}
}
