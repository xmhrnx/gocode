package settingsview

import (
	"fmt"

	"gocode/internal/settings"
	"gocode/ui/core"
)

func appearanceEntries(s *core.Services) []entry {
	a := &s.Settings.Appearance
	points := func(v int) string { return fmt.Sprintf("%d pt", v) }
	width := func(v int) string {
		if v >= 10000 {
			return "Full window"
		}
		return fmt.Sprintf("%d dp", v)
	}
	barStyles := []option[string]{{"windows", "Buttons on the right"}, {"mac", "Buttons on the left"}}
	fps := func(v int) string {
		if v == settings.NoFrameCap {
			return "No cap"
		}
		return fmt.Sprintf("%d fps", v)
	}
	return []entry{
		{title: "Theme", description: "Pick a look; System follows your OS: Obsidian when it is dark, Sand when it is light.",
			keywords: "theme color dark light palette appearance", wide: true, control: newThemeGrid(s)},
		{title: "Text size", description: "Size of the interface and chat text.",
			keywords: "font size text scale zoom", control: intChoice(s, &a.UIFontSize, []int{12, 13, 14, 15, 16, 17, 18}, points)},
		{title: "Code size", description: "Size of code, commands and tool output.",
			keywords: "font size code monospace", control: intChoice(s, &a.CodeFontSize, []int{11, 12, 13, 14, 15, 16}, points)},
		{title: "Conversation width", description: "Widest the chat column grows on large windows.",
			keywords: "width column layout", control: intChoice(s, &a.ChatWidth, []int{680, 820, 960, 1120, 10000}, width)},
		{title: "Frame rate cap", description: "Most frames a second while something moves; lower saves power. An idle window draws nothing either way.",
			keywords: "fps frame rate refresh performance battery smooth", control: intChoice(s, &a.MaxFPS, []int{30, 60, 120, 144, 240, settings.NoFrameCap}, fps)},
		{title: "Window buttons", description: "Where minimize, maximize and close sit in the title bar.",
			keywords: "title bar window controls mac windows", control: choice(s, &a.TitleBarStyle, barStyles)},
	}
}
