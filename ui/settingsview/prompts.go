package settingsview

import (
	"github.com/xmhrnx/giode"

	"gocode/internal/prompts"
	"gocode/ui/core"
	"gocode/ui/theme"
)

func promptEntries(s *core.Services) []entry {
	p := &s.Settings.Prompts
	return []entry{
		{title: "Your rules", description: "Added to every agent prompt: coding style, tools to prefer, things to avoid.",
			keywords: "custom instructions rules system prompt", wide: true,
			control: promptEditor(s, &p.CustomRules, "", "e.g. Use pnpm, not npm. Prefer table-driven tests.")},
		{title: "Agent prompt", description: "The system prompt for chats with tools. Placeholders: {{workdir}}, {{os}}, {{shell}}, {{date}}.",
			keywords: "system prompt agent instructions", wide: true,
			control: promptEditor(s, &p.Agent, prompts.DefaultAgent(), "")},
		{title: "Chat prompt", description: "The system prompt for chats without tools.",
			keywords: "system prompt chat no tools", wide: true,
			control: promptEditor(s, &p.Chat, prompts.DefaultChat(), "")},
		{title: "Compaction prompt", description: "Asks the model to summarize a conversation when its context is compacted.",
			keywords: "compact compaction summary prompt context", wide: true,
			control: promptEditor(s, &p.Compact, prompts.DefaultCompact(), "")},
		{title: "Title prompt", description: "Asks the model to name a chat after its first message.",
			keywords: "title prompt naming", wide: true,
			control: promptEditor(s, &p.Title, prompts.DefaultTitle(), "")},
	}
}

// promptEditor edits a prompt that falls back to a built-in default. The
// field shows the default until changed; text equal to the default is
// stored as empty, so future improvements to the default still apply.
func promptEditor(s *core.Services, value *string, fallback, placeholder string) giode.Element {
	field := giode.Input.Text(placeholder, giode.Styles{Width: -1, Background: theme.Canvas, BorderColor: theme.Border,
		BorderRadius: 8, Color: theme.Text, FontSize: theme.Sizes.Small}).Lines(4, 18)
	field.SetText(prompts.Or(*value, fallback))
	field.OnChange(func(text string) {
		if text == fallback {
			text = ""
		}
		*value = text
		s.SettingsChanged()
	})
	controls := []giode.Element{field}
	if fallback != "" {
		reset := giode.Button("Reset to default", giode.Styles{Color: theme.Muted, Background: theme.Hover,
			FontSize: theme.Sizes.Small, BorderRadius: 7, Padding: giode.SymmetricInset(10, 5)}).Variant(giode.VariantGhost)
		reset.OnClick(func() { field.SetText(fallback); *value = ""; s.SettingsChanged() })
		controls = append(controls, reset)
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 6, Align: giode.AlignEnd}, controls...)
}
