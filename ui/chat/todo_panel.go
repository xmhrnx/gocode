package chat

import (
	"github.com/xmhrnx/giode"

	"gocode/internal/tools"
	"gocode/ui/components"
	"gocode/ui/theme"
)

// todoPanel shows the agent's task list above the composer while some of
// it is unfinished. Closing it hides the list until the agent writes a
// new one.
type todoPanel struct {
	tab   *ChatTab
	close *components.IconButton
}

func newTodoPanel(tab *ChatTab) *todoPanel {
	p := &todoPanel{tab: tab}
	p.close = components.NewIconButton("x", "Hide the plan", func() {
		tab.todos = nil
		tab.changed()
	}).SetSize(13)
	return p
}

func (p *todoPanel) Layout(gtx giode.Context) giode.Dimensions {
	todos := p.tab.todos
	if !hasOpenTodos(todos) {
		return giode.Dimensions{}
	}
	header := giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
		giode.Text("Plan", giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Tiny, FlexGrow: 1,
			FontWeight: giode.FontWeightSemiBold, TextTransform: giode.TextTransformUppercase}),
		p.close)
	rows := []giode.Element{header}
	for _, todo := range todos {
		rows = append(rows, todoRow(todo))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 5, Padding: giode.Inset{Left: 14, Right: 8, Top: 6, Bottom: 10}, BorderRadius: 12,
		Background: theme.Panel, BorderWidth: 1, BorderColor: theme.SoftBorder}, rows...).Layout(gtx)
}

func todoRow(todo tools.Todo) giode.Element {
	icon, color, text := "circle", theme.Faint, theme.Muted
	switch todo.Status {
	case "in_progress":
		icon, color, text = "circle-dot", theme.Primary, theme.Text
	case "completed":
		icon, color = "circle-check", theme.Success
	}
	style := giode.Styles{Color: text, FontSize: theme.Sizes.Small}
	if todo.Status == "completed" {
		style.TextDecoration = giode.TextDecorationLineThrough
	}
	return giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
		theme.Icon(icon, 13, color), giode.Text(todo.Content, style))
}

func hasOpenTodos(todos []tools.Todo) bool {
	for _, todo := range todos {
		if todo.Status != "completed" {
			return true
		}
	}
	return false
}
