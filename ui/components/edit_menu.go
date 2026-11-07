package components

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"

	"gocode/ui/theme"
)

// EditOptions adjusts a field's edit menu.
type EditOptions struct {
	// Secret leaves out Cut and Copy, for keys and passwords.
	Secret bool
	// Paste replaces the plain text paste; the composer uses it to
	// attach a copied image.
	Paste func()
}

// AddEditMenu gives field a right-click (or long-press) menu with Cut,
// Copy, Paste and Select all, opening at the pointer.
func AddEditMenu(app *giode.App, field *input.Input, opts EditOptions) {
	paste := opts.Paste
	if paste == nil {
		paste = func() { PasteText(app, field) }
	}
	m := giode.Menu(theme.MenuStyle())
	if !opts.Secret {
		m.Item("Cut", func() { cut(app, field) }).
			Item("Copy", func() { copySelection(app, field) })
	}
	m.Item("Paste", paste).
		Separator().
		Item("Select all", func() { field.SelectAll(); field.Focus(); redraw(app) })
	field.Context(m)
}

// PasteText puts the clipboard's text in place of field's selection.
func PasteText(app *giode.App, field *input.Input) {
	app.ReadClipboard(func(text string) {
		if text != "" {
			field.InsertText(text)
		}
		field.Focus()
		redraw(app)
	})
}

func cut(app *giode.App, field *input.Input) {
	if copySelection(app, field) {
		field.InsertText("")
		redraw(app)
	}
}

// copySelection copies field's selected text and reports whether there
// was any.
func copySelection(app *giode.App, field *input.Input) bool {
	defer field.Focus()
	start, end := field.Selection()
	runes := []rune(field.Text())
	if start >= end || end > len(runes) {
		return false
	}
	app.SetClipboard(string(runes[start:end]))
	return true
}

// redraw shows a change made to a field from code: cached drawings can't
// see it, so they are dropped and a frame is drawn.
func redraw(app *giode.App) {
	giode.InvalidateCaches()
	app.Invalidate()
}
