package ui

import (
	"gioui.org/layout"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/dialog"

	"gocode/ui/theme"
)

// closeConfirmation asks before quitting while the agent is working.
type closeConfirmation struct {
	open   bool
	dialog *dialog.Dialog
	cancel *giode.ButtonElement
	quit   *giode.ButtonElement
}

func newCloseConfirmation(app *giode.App) *closeConfirmation {
	c := &closeConfirmation{}
	c.cancel = giode.Button("Keep working", giode.Styles{BorderColor: theme.Border, Color: theme.Text, BorderRadius: 8}).
		Variant(giode.VariantOutlined).OnClick(func() { c.open = false })
	c.quit = giode.Button("Quit", giode.Styles{Background: theme.Danger, Color: "#ffffff", BorderRadius: 8}).
		OnClick(app.Close)
	c.dialog = giode.Dialog(&c.open, c.content, giode.Styles{Background: theme.Panel, BorderRadius: 14,
		BorderWidth: 1, BorderColor: theme.Border})
	return c
}

// allow reports whether the window may close now; when work is running
// it asks first and returns false.
func (c *closeConfirmation) allow(busy bool) bool {
	if !busy {
		return true
	}
	c.open = true
	return false
}

func (c *closeConfirmation) Layout(gtx layout.Context) layout.Dimensions { return c.dialog.Layout(gtx) }

func (c *closeConfirmation) content() giode.Element {
	return giode.Stack(giode.Styles{Width: 380, Gap: 12},
		giode.Text("The agent is still working", giode.Styles{Color: theme.Text, FontSize: theme.Sizes.Title, FontWeight: giode.FontWeightSemiBold}),
		giode.Text("Quitting stops it. Finished steps are saved in the chat history.", giode.Styles{Color: theme.Muted}),
		giode.Row(giode.Styles{Width: -1, Justify: giode.JustifyEnd, Gap: 8, Margin: giode.Inset{Top: 8}}, c.cancel, c.quit),
	)
}
