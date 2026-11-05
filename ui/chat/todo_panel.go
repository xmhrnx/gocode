package chat

import (
	"github.com/xmhrnx/giode"

	"gocode/internal/tools"
	"gocode/ui/theme"
)

// todoPanel shows the agent's task list above the composer while some of
// it is unfinished.
func todoPanel(c *ChatTab) giode.Element {
	if !hasOpenTodos(c.todos) {
		return giode.Box(giode.Styles{Display: giode.DisplayNone})
	}
	rows := []giode.Element{giode.Text("Plan", giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Tiny,
		FontWeight: giode.FontWeightSemiBold, TextTransform: giode.TextTransformUppercase})}
	for _, todo := range c.todos {
		rows = append(rows, todoRow(todo))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 5, Padding: giode.SymmetricInset(14, 10), BorderRadius: 12,
		Background: theme.Panel, BorderWidth: 1, BorderColor: theme.SoftBorder}, rows...)
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
