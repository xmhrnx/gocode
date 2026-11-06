package settingsview

import (
	"github.com/xmhrnx/giode"

	"gocode/ui/components"
	"gocode/ui/core"
	"gocode/ui/modelpicker"
)

// modelChoice binds a model setting ("provider/model") to a button that
// opens the model picker.
type modelChoice struct {
	value  *string
	empty  string
	pill   *components.PillButton
	picker *modelpicker.Picker
}

func newModelChoice(s *core.Services, value *string, empty string) *modelChoice {
	m := &modelChoice{value: value, empty: empty}
	m.picker = modelpicker.New(s, func(ref string) { *value = ref; s.SettingsChanged() })
	m.picker.Current = func() string { return *value }
	m.pill = components.NewPillButton("cpu", "Choose a model", m.picker.Open)
	return m
}

func (m *modelChoice) Layout(gtx giode.Context) giode.Dimensions {
	label := *m.value
	if label == "" {
		label = m.empty
	}
	m.pill.SetLabel(label)
	return giode.Stack(giode.Styles{}, m.pill, m.picker).Layout(gtx)
}
