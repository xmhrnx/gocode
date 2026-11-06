package settingsview

import (
	"github.com/xmhrnx/giode"

	"gocode/ui/core"
	"gocode/ui/theme"
)

func aboutEntries(s *core.Services) []entry {
	version := giode.AppVersion
	if version == "" {
		version = "development build"
	}
	return []entry{
		{title: "GoCode", description: "A native coding agent and terminal, built with Go and Giode. No browser engine inside.",
			keywords: "version about", control: giode.Text(version, theme.Caption())},
		{title: "Settings file", description: "Settings are plain JSON. API keys are kept in the system keychain, not here.",
			keywords: "config path file json location", control: giode.Text(s.SettingsPath, giode.Styles{
				Color: theme.Muted, FontSize: theme.Sizes.Small, FontFamily: "JetBrains Mono"}).Selectable("settings-path")},
	}
}
