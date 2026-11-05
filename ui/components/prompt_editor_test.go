package components

import (
	"testing"

	"gioui.org/io/key"

	"gocode/internal/settings"
)

func TestIsSendFollowsTheSendKey(t *testing.T) {
	cases := []struct {
		sendKey settings.SendKey
		mods    key.Modifiers
		want    bool
	}{
		{settings.SendEnter, 0, true},
		{settings.SendEnter, key.ModShift, false},
		{settings.SendShiftEnter, 0, false},
		{settings.SendShiftEnter, key.ModShift, true},
		{settings.SendCtrlEnter, 0, false},
		{settings.SendCtrlEnter, key.ModShift, false},
		{settings.SendCtrlEnter, key.ModShortcut, true},
	}
	for _, c := range cases {
		editor := NewPromptEditor("", func() settings.SendKey { return c.sendKey }, func() {})
		if got := editor.isSend(c.mods); got != c.want {
			t.Errorf("send key %s with modifiers %v: got %v, want %v", c.sendKey, c.mods, got, c.want)
		}
	}
}
