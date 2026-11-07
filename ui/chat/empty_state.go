package chat

import (
	"github.com/xmhrnx/giode"

	"gocode/ui/theme"
)

// emptyState greets a new chat with starter prompts.
type emptyState struct {
	tab      *ChatTab
	starters []giode.Element
}

var starterPrompts = []struct{ icon, label, prompt string }{
	{"book-open", "Explain this codebase", "Give me a concise tour of this codebase: its purpose, structure, main entry points and how the pieces fit together."},
	{"search", "Find and fix a bug", "There's a bug: "},
	{"circle-check", "Write tests", "Write tests for "},
	{"git-branch", "Review my changes", "Review the uncommitted changes in this repository (git diff) for bugs and risky code."},
}

func newEmptyState(tab *ChatTab) *emptyState {
	e := &emptyState{tab: tab}
	for _, s := range starterPrompts {
		e.starters = append(e.starters, giode.Pressable(func(state giode.PressState) giode.Element {
			return starterChip(s.icon, s.label, state.Hovered)
		}).Label(s.label).OnClick(func(giode.ClickEvent) {
			tab.composer.editor.SetText(s.prompt)
			tab.composer.editor.Focus()
		}))
	}
	return e
}

func (e *emptyState) Layout(gtx giode.Context) giode.Dimensions {
	return giode.Stack(giode.Styles{Width: -1, Height: -1, Align: giode.AlignCenter, Justify: giode.JustifyCenter, Gap: 14,
		Padding: giode.UniformInset(24)},
		giode.Box(giode.Styles{Width: 52, Height: 52, BorderRadius: 16, Background: theme.PrimarySoft,
			Align: giode.AlignCenter, Justify: giode.JustifyCenter}, theme.Icon("square-terminal", 26, theme.Primary)),
		giode.Text("What are we building?", giode.Styles{Color: theme.Text, FontSize: theme.Sizes.Heading, FontWeight: giode.FontWeightSemiBold}),
		giode.Text(e.tab.folderLabel(), giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Small, FontFamily: "JetBrains Mono"}),
		giode.Row(giode.Styles{Gap: 8, Justify: giode.JustifyCenter, MaxWidth: 760, Margin: giode.Inset{Top: 10}}, e.starters...),
	).Layout(gtx)
}

func starterChip(icon, label string, hovered bool) giode.Element {
	background, color := theme.Panel, theme.Muted
	if hovered {
		background, color = theme.Hover, theme.Text
	}
	return giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter, Padding: giode.SymmetricInset(12, 8),
		BorderRadius: 10, Background: background, BorderWidth: 1, BorderColor: theme.SoftBorder},
		theme.Icon(icon, 14, color),
		giode.Text(label, giode.Styles{Color: color, FontSize: theme.Sizes.Small}))
}
