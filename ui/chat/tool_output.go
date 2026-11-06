package chat

import (
	"fmt"
	"strings"

	"github.com/xmhrnx/giode"

	"gocode/ui/theme"
)

// toolOutput shows a tool's text output in a monospace block.
func toolOutput(text string, isError bool) giode.Element {
	shown, hidden := clipLines(text, maxShownLines)
	style := theme.Mono()
	style.FontSize = theme.Sizes.Code - 1
	if isError {
		style.Color = theme.Danger
	}
	children := []giode.Element{giode.Text(shown, style)}
	if hidden > 0 {
		children = append(children, hiddenLinesNote(hidden))
	}
	return giode.Stack(giode.Styles{Width: -1, Padding: giode.SymmetricInset(14, 10), Gap: 6}, children...)
}

// diffView shows a unified diff with added and removed lines colored.
func diffView(diff string) giode.Element {
	shown, hidden := clipLines(diff, maxShownLines)
	var lines []giode.Element
	for _, line := range strings.Split(shown, "\n") {
		lines = append(lines, diffLine(line))
	}
	if hidden > 0 {
		lines = append(lines, hiddenLinesNote(hidden))
	}
	return giode.Stack(giode.Styles{Width: -1, Padding: giode.SymmetricInset(0, 8)}, lines...)
}

func diffLine(line string) giode.Element {
	style := theme.Mono()
	style.FontSize = theme.Sizes.Code - 1
	background := ""
	switch {
	case strings.HasPrefix(line, "+++"), strings.HasPrefix(line, "---"):
		style.Color = theme.Muted
	case strings.HasPrefix(line, "@@"):
		style.Color = theme.Primary
	case strings.HasPrefix(line, "+"):
		style.Color, background = theme.Success, theme.SuccessSoft
	case strings.HasPrefix(line, "-"):
		style.Color, background = theme.Danger, theme.DangerSoft
	}
	if line == "" {
		line = " "
	}
	return giode.Box(giode.Styles{Width: -1, Background: background, Padding: giode.SymmetricInset(14, 0)},
		giode.Text(line, style))
}

func hiddenLinesNote(hidden int) giode.Element {
	return giode.Box(giode.Styles{Padding: giode.Inset{Left: 14}},
		giode.Text(fmt.Sprintf("… %d more lines", hidden), giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Tiny}))
}
