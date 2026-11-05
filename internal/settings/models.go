package settings

import "slices"

const maxRecentModels = 8

// MarkUsed moves model ("provider/model") to the front of the recent list.
func (m *Models) MarkUsed(model string) {
	m.Recent = slices.DeleteFunc(m.Recent, func(id string) bool { return id == model })
	m.Recent = append([]string{model}, m.Recent...)
	if len(m.Recent) > maxRecentModels {
		m.Recent = m.Recent[:maxRecentModels]
	}
}

// IsHidden reports whether the model is hidden from the picker.
func (m *Models) IsHidden(model string) bool {
	return slices.Contains(m.Hidden, model)
}

// SetHidden hides or shows the model in the picker.
func (m *Models) SetHidden(model string, hidden bool) {
	m.Hidden = slices.DeleteFunc(m.Hidden, func(id string) bool { return id == model })
	if hidden {
		m.Hidden = append(m.Hidden, model)
	}
}
