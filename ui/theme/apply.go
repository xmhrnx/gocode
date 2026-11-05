package theme

import "github.com/xmhrnx/giode"

// Apply switches the app to the theme with the given ID. SystemID follows
// the operating system's dark mode setting.
func Apply(id string) {
	t := Resolve(id)
	mode := giode.ModeLight
	if t.Dark {
		mode = giode.ModeDark
	}
	giode.SetPalette(mode, t.Palette)
	giode.SetMode(mode)
}

// Resolve returns the theme an ID stands for, resolving SystemID.
func Resolve(id string) Theme {
	if id != SystemID {
		return Find(id)
	}
	if dark, err := giode.SystemDarkMode(); err == nil && !dark {
		return Find("paper")
	}
	return Find("midnight")
}
