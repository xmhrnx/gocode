package chat

import (
	"gioui.org/layout"
	"github.com/xmhrnx/giode"

	"gocode/internal/llm"
	"gocode/ui/theme"
)

// toolRow shows one tool call: what it does, its status, and, expanded,
// its output or diff.
type toolRow struct {
	rowBase
	expanded bool
	header   *giode.PressableElement
	copy     giode.Element
}

func newToolRow(tab *ChatTab, it item) *toolRow {
	r := &toolRow{rowBase: rowBase{tab: tab, it: it}, expanded: tab.services.Settings.Chat.ExpandToolOutput}
	r.header = giode.Pressable(r.headerContent).Label("Toggle tool output").
		OnClick(func(giode.ClickEvent) { r.expanded = !r.expanded })
	r.copy = giode.CopyButton(func() string {
		if result := tab.toolResult(r.call().ID); result != nil {
			return result.Content
		}
		return ""
	}, giode.Styles{Color: theme.Muted})
	return r
}

func (r *toolRow) version() int {
	if r.expanded {
		return 1
	}
	return 0
}

func (r *toolRow) call() llm.ToolCall {
	part, ok := r.tab.part(r.it)
	if !ok || part.ToolCall == nil {
		return llm.ToolCall{}
	}
	return *part.ToolCall
}

func (r *toolRow) Layout(gtx layout.Context) layout.Dimensions {
	children := []giode.Element{r.header}
	result := r.tab.toolResult(r.call().ID)
	if r.expanded && result != nil {
		children = append(children, giode.Divider(giode.Styles{Color: theme.SoftBorder}), r.body(result))
	}
	return giode.Box(giode.Styles{Width: -1, Padding: giode.SymmetricInset(0, 3)},
		giode.Stack(giode.Styles{Width: -1, BorderRadius: 10, BorderWidth: 1, BorderColor: theme.SoftBorder,
			Background: theme.Panel, Overflow: giode.OverflowHidden}, children...),
	).Layout(gtx)
}

func (r *toolRow) headerContent(state giode.PressState) giode.Element {
	call := r.call()
	result := r.tab.toolResult(call.ID)
	background := ""
	if state.Hovered && result != nil {
		background = theme.Hover
	}
	chevron := giode.Element(giode.Box(giode.Styles{}))
	if result != nil {
		chevron = theme.Icon(map[bool]string{true: "chevron-down", false: "chevron-right"}[r.expanded], 13, theme.Muted)
	}
	return giode.Row(giode.Styles{Width: -1, Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
		Padding: giode.SymmetricInset(12, 8), Background: background},
		r.statusIcon(call.ID, result),
		theme.Icon(toolIcon(call.Name), 14, theme.Muted),
		giode.Text(r.tab.toolSummary(call.Name, call.Arguments), giode.Styles{
			Color: theme.Text, FontSize: theme.Sizes.Small, FontFamily: "JetBrains Mono", FlexGrow: 1,
			TextOverflow: giode.TextOverflowEllipsis}),
		chevron,
	)
}

func (r *toolRow) statusIcon(callID string, result *llm.ToolResult) giode.Element {
	switch {
	case result != nil && result.IsError:
		return theme.Icon("circle-x", 14, theme.Danger)
	case result != nil:
		return theme.Icon("circle-check", 14, theme.Success)
	case r.tab.toolRunning(callID):
		return giode.Spinner(giode.Styles{Width: 14, Height: 14, Color: theme.Primary})
	}
	return theme.Icon("clock", 14, theme.Muted)
}

func (r *toolRow) body(result *llm.ToolResult) giode.Element {
	content := toolOutput(r.it.key, result.Content, result.IsError)
	if display := r.tab.session.ToolDisplays[r.call().ID]; display != "" {
		content = diffView(display)
	}
	return giode.Stack(giode.Styles{Width: -1, Background: theme.CodeBack},
		content,
		giode.Row(giode.Styles{Width: -1, Justify: giode.JustifyEnd, Padding: giode.Inset{Right: 6, Bottom: 4}}, r.copy),
	)
}
