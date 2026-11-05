package components

import (
	"strings"

	"gioui.org/io/key"
	"gioui.org/layout"

	"gocode/internal/settings"
)

// handleEnter takes every Enter press before the editor sees it, then
// either sends or inserts a newline, depending on the send key.
func (p *PromptEditor) handleEnter(gtx layout.Context) {
	mods := key.ModShift | key.ModShortcut
	for {
		ev, ok := gtx.Event(
			key.Filter{Focus: &p.editor, Name: key.NameReturn, Optional: mods},
			key.Filter{Focus: &p.editor, Name: key.NameEnter, Optional: mods},
		)
		if !ok {
			return
		}
		press, ok := ev.(key.Event)
		if !ok || press.State != key.Press {
			continue
		}
		if p.isSend(press.Modifiers) {
			p.onSend()
		} else if p.editor.Insert("\n") > 0 && p.onChange != nil {
			p.onChange()
		}
	}
}

// handleArrows takes Up and Down from the editor while there is only one
// line, where they would have nothing to move through.
func (p *PromptEditor) handleArrows(gtx layout.Context) {
	if p.onArrow == nil || strings.Contains(p.editor.Text(), "\n") {
		return
	}
	for {
		ev, ok := gtx.Event(
			key.Filter{Focus: &p.editor, Name: key.NameUpArrow},
			key.Filter{Focus: &p.editor, Name: key.NameDownArrow},
		)
		if !ok {
			return
		}
		if press, ok := ev.(key.Event); ok && press.State == key.Press {
			p.onArrow(press.Name == key.NameUpArrow)
		}
	}
}

func (p *PromptEditor) isSend(mods key.Modifiers) bool {
	shift, ctrl := mods.Contain(key.ModShift), mods.Contain(key.ModShortcut)
	switch p.sendKey() {
	case settings.SendShiftEnter:
		return shift && !ctrl
	case settings.SendCtrlEnter:
		return ctrl
	default:
		return !shift && !ctrl
	}
}
