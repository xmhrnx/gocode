package settingsview

import (
	"strconv"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/styles/properties"

	"gocode/ui/core"
	"gocode/ui/theme"

	"gocode/ui/components"
)

// option is one choice of a string-valued setting.
type option[T ~string] struct {
	value T
	label string
}

// switchControl binds a switch to a boolean setting.
func switchControl(s *core.Services, value *bool) giode.Element {
	sw := giode.Input.Switch("")
	sw.SetChecked(*value)
	sw.OnChange(func(on bool) { *value = on; s.SettingsChanged() })
	return sw
}

// choice binds a dropdown to a string-valued setting.
func choice[T ~string](s *core.Services, value *T, options []option[T]) giode.Element {
	labels := make([]string, len(options))
	selected := 0
	for i, o := range options {
		labels[i] = o.label
		if o.value == *value {
			selected = i
		}
	}
	sel := giode.Input.Select(fieldStyle(200), labels...)
	sel.SetSelected(selected)
	sel.OnChange(func(i int) { *value = options[i].value; s.SettingsChanged() })
	return sel
}

// number binds a text field to an integer setting within [low, high];
// text that isn't a number in range is ignored until it is.
func number(s *core.Services, value *int, low, high int) giode.Element {
	field := giode.Input.Text(strconv.Itoa(*value), fieldStyle(110)).Keyboard(giode.KeyboardNumeric)
	components.AddEditMenu(s.App, field, components.EditOptions{})
	field.SetText(strconv.Itoa(*value))
	field.OnChange(func(text string) {
		if n, err := strconv.Atoi(text); err == nil && n >= low && n <= high {
			*value = n
			s.SettingsChanged()
		}
	})
	return field
}

// decimal binds a text field to a float setting within [low, high].
func decimal(s *core.Services, value *float64, low, high float64) giode.Element {
	text := strconv.FormatFloat(*value, 'f', -1, 64)
	field := giode.Input.Text(text, fieldStyle(110)).Keyboard(giode.KeyboardNumeric)
	components.AddEditMenu(s.App, field, components.EditOptions{})
	field.SetText(text)
	field.OnChange(func(text string) {
		if n, err := strconv.ParseFloat(text, 64); err == nil && n >= low && n <= high {
			*value = n
			s.SettingsChanged()
		}
	})
	return field
}

// text binds a single-line field to a string setting.
func text(s *core.Services, value *string, placeholder string, width int) giode.Element {
	field := giode.Input.Text(placeholder, fieldStyle(width))
	components.AddEditMenu(s.App, field, components.EditOptions{})
	field.SetText(*value)
	field.OnChange(func(text string) { *value = text; s.SettingsChanged() })
	return field
}

func fieldStyle(width int) giode.Styles {
	return giode.Styles{Width: properties.Width(width), Background: theme.Canvas, BorderColor: theme.Border,
		BorderRadius: 8, Color: theme.Text, FontSize: theme.Sizes.Small}
}

// intChoice binds a dropdown to an integer setting with fixed values.
func intChoice(s *core.Services, value *int, values []int, label func(int) string) giode.Element {
	labels := make([]string, len(values))
	selected := 0
	for i, v := range values {
		labels[i] = label(v)
		if v == *value {
			selected = i
		}
	}
	sel := giode.Input.Select(fieldStyle(200), labels...)
	sel.SetSelected(selected)
	sel.OnChange(func(i int) { *value = values[i]; s.SettingsChanged() })
	return sel
}
