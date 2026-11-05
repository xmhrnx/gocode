package components

import (
	"image"

	"gioui.org/font"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/elements"
	"github.com/xmhrnx/giode/fonts"
	"github.com/xmhrnx/giode/styles/properties"

	"gocode/internal/settings"
	"gocode/ui/theme"
)

// PromptEditor is a multi-line text field that grows with its content and
// sends on a configurable key: Enter, Shift+Enter or Ctrl+Enter. The other
// Enter combinations insert a newline. It is built on Gio's editor
// directly because a field's own Enter handling can't be redefined.
type PromptEditor struct {
	editor      widget.Editor
	placeholder string
	sendKey     func() settings.SendKey
	onSend      func()
	onChange    func()
	onArrow     func(up bool)
	maxLines    int
	monospace   bool
	// tap is the tag of the press area that focuses the editor.
	tap       bool
	focusNext bool
}

// NewPromptEditor returns an editor that calls onSend when sendKey (read
// every frame, so settings changes apply at once) is pressed.
func NewPromptEditor(placeholder string, sendKey func() settings.SendKey, onSend func()) *PromptEditor {
	return &PromptEditor{placeholder: placeholder, sendKey: sendKey, onSend: onSend, maxLines: 14}
}

// OnChange sets a handler called after every edit.
func (p *PromptEditor) OnChange(fn func()) *PromptEditor { p.onChange = fn; return p }

// OnArrow sets a handler for Up and Down while the content is a single
// line, for recalling earlier input.
func (p *PromptEditor) OnArrow(fn func(up bool)) *PromptEditor { p.onArrow = fn; return p }

// Monospace shows the text in the code font, for commands.
func (p *PromptEditor) Monospace() *PromptEditor { p.monospace = true; return p }

// Text returns the content.
func (p *PromptEditor) Text() string { return p.editor.Text() }

// SetText replaces the content and puts the caret at the end.
func (p *PromptEditor) SetText(text string) {
	p.editor.SetText(text)
	p.editor.SetCaret(p.editor.Len(), p.editor.Len())
}

// Focus requests keyboard focus on the next frame.
func (p *PromptEditor) Focus() { p.focusNext = true }

// Layout lays the editor out.
func (p *PromptEditor) Layout(gtx layout.Context) layout.Dimensions {
	if p.focusNext || p.pressed(gtx) {
		p.focusNext = false
		gtx.Execute(key.FocusCmd{Tag: &p.editor})
	}
	p.handleEnter(gtx)
	p.handleArrows(gtx)
	for {
		ev, ok := p.editor.Update(gtx)
		if !ok {
			break
		}
		if _, changed := ev.(widget.ChangeEvent); changed && p.onChange != nil {
			p.onChange()
		}
	}
	size, face := theme.Sizes.Base, font.Font{Typeface: "Inter"}
	if p.monospace {
		size, face = theme.Sizes.Code, font.Font{Typeface: "JetBrains Mono"}
	}
	lineHeight := gtx.Sp(size * 1.5)
	gtx.Constraints.Min.Y = lineHeight
	gtx.Constraints.Max.Y = min(gtx.Constraints.Max.Y, lineHeight*p.maxLines)
	p.editor.LineHeightScale = 1.5
	if p.editor.Len() == 0 {
		p.layoutPlaceholder(gtx, size, face)
	}
	dims := p.editor.Layout(gtx, fonts.Shaper(), face, size, colorMaterial(gtx, theme.Text), colorMaterial(gtx, theme.PrimarySoft))
	p.registerTap(gtx, dims.Size)
	elements.RecordSemantic(gtx, elements.SemanticNode{Kind: elements.SemanticInput, Label: p.placeholder,
		Value: p.editor.Text(), Target: p}, dims.Size)
	return dims
}

// ChangeText replaces the content as typing would; with SubmitText it
// lets automation (Giode's MCP server) drive the editor.
func (p *PromptEditor) ChangeText(text string) {
	p.SetText(text)
	if p.onChange != nil {
		p.onChange()
	}
}

// SubmitText sends the content, as the send key would.
func (p *PromptEditor) SubmitText() { p.onSend() }

func (p *PromptEditor) layoutPlaceholder(gtx layout.Context, size unit.Sp, face font.Font) {
	giode.Text(p.placeholder, giode.Styles{Color: theme.Faint, FontSize: size, MaxLines: 1, FontFamily: properties.FontFamily(face.Typeface),
		TextOverflow: giode.TextOverflowEllipsis, LineHeight: properties.LineHeight(size * 1.5)}).Layout(gtx)
}

func colorMaterial(gtx layout.Context, color string) op.CallOp {
	macro := op.Record(gtx.Ops)
	paint.ColorOp{Color: properties.CalcColor(color)}.Add(gtx.Ops)
	return macro.Stop()
}

// pressed reports a press on the editor. The editor focuses itself on
// click too, but after a popup menu closes its own click handling can miss
// the press; an explicit focus on press, as Giode's text field does, makes
// clicking the field always work.
func (p *PromptEditor) pressed(gtx layout.Context) bool {
	pressed := false
	for {
		ev, ok := gtx.Event(pointer.Filter{Target: &p.tap, Kinds: pointer.Press})
		if !ok {
			return pressed
		}
		if _, isPointer := ev.(pointer.Event); isPointer {
			pressed = true
		}
	}
}

// registerTap listens for presses over the editor, letting them pass on
// to the editor beneath for caret placement and selection.
func (p *PromptEditor) registerTap(gtx layout.Context, size image.Point) {
	pass := pointer.PassOp{}.Push(gtx.Ops)
	area := clip.Rect{Max: size}.Push(gtx.Ops)
	event.Op(gtx.Ops, &p.tap)
	area.Pop()
	pass.Pop()
}
