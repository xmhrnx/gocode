package settingsview

import (
	"fmt"
	"runtime"
	"time"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/kit"

	"gocode/ui/core"
	"gocode/ui/theme"
)

// memoryLabel shows the Go runtime's memory use, refreshed every few
// seconds while visible.
type memoryLabel struct {
	text    string
	updated time.Time
}

func aboutEntries(s *core.Services) []entry {
	version := giode.AppVersion
	if version == "" {
		version = "development build"
	}
	return []entry{
		{title: "GoCode", description: "A native coding agent and terminal, built with Go and Giode. No browser engine inside.",
			keywords: "version about", control: giode.Text(version, theme.Caption())},
		{title: "Memory", description: "Memory the Go runtime holds right now, updated every few seconds.",
			keywords: "ram memory usage performance", control: &memoryLabel{}},
		{title: "Settings file", description: "Settings are plain JSON. API keys are kept in the system keychain, not here.",
			keywords: "config path file json location", control: giode.Text(s.SettingsPath, giode.Styles{
				Color: theme.Muted, FontSize: theme.Sizes.Small, FontFamily: "JetBrains Mono"}).Selectable("settings-path")},
	}
}

func (m *memoryLabel) Layout(gtx giode.Context) giode.Dimensions {
	if time.Since(m.updated) > 3*time.Second {
		var stats runtime.MemStats
		runtime.ReadMemStats(&stats)
		m.text = fmt.Sprintf("%.1f MB in use · %.1f MB reserved", float64(stats.HeapInuse)/(1<<20), float64(stats.Sys)/(1<<20))
		m.updated = time.Now()
	}
	kit.RequestFrameAt(gtx, m.updated.Add(3*time.Second))
	return giode.Text(m.text, giode.Styles{Color: theme.Text, FontSize: theme.Sizes.Small, FontFamily: "JetBrains Mono"}).Layout(gtx)
}
