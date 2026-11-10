// Package theme holds GoCode's color themes, typography and shared styles.
package theme

import "github.com/xmhrnx/giode"

// Theme is one selectable look.
type Theme struct {
	ID      string
	Name    string
	Dark    bool
	Palette giode.Palette
}

// SystemID picks a dark or light theme following the operating system:
// DefaultDark or DefaultLight.
const SystemID = "system"

// The themes System uses in dark and light mode.
const (
	DefaultDark  = "obsidian"
	DefaultLight = "sand"
)

// All returns the built-in themes, dark ones first.
func All() []Theme { return themes }

// Find returns the theme with the given ID, or the default dark theme.
func Find(id string) Theme {
	for _, t := range themes {
		if t.ID == id {
			return t
		}
	}
	for _, t := range themes {
		if t.ID == DefaultDark {
			return t
		}
	}
	return themes[0]
}

var themes = []Theme{
	dark("midnight", "Midnight", "#0b0d12", "#12151c", "#e6e8ee", "#232836", "#8b93a7", "#7c8cff", "#ffffff", "#b18cff"),
	dark("obsidian", "Obsidian", "#09090a", "#121213", "#ededed", "#262628", "#8f8f94", "#f5a524", "#1a1200", "#ffcf70"),
	dark("forest", "Forest", "#0a100d", "#101915", "#e4ece7", "#1f2c25", "#86a094", "#3ecf8e", "#04140c", "#7ee0b4"),
	dark("ember", "Ember", "#110d0b", "#1a1411", "#f0e7e1", "#2e231d", "#a8968b", "#ff7a45", "#1a0b04", "#ffb38f"),
	dark("fjord", "Fjord", "#1b1f27", "#222733", "#e5e9f0", "#323a4a", "#8a94a8", "#88c0d0", "#10151c", "#b48ead"),
	dark("orchid", "Orchid", "#120d12", "#1a1319", "#f0e6ee", "#2c2029", "#a38f9e", "#f472b6", "#1a0612", "#c4a1ff"),
	light("paper", "Paper", "#f7f7f5", "#ffffff", "#1c1d21", "#e4e4e0", "#6b6f78", "#4f5bd5", "#ffffff", "#8b5cf6"),
	light("daylight", "Daylight", "#f3f6fb", "#ffffff", "#0f172a", "#dbe2ee", "#64748b", "#2563eb", "#ffffff", "#0891b2"),
	light("sand", "Sand", "#f4efe6", "#fffaf2", "#2b2520", "#e4d9c8", "#7c6f62", "#c2410c", "#ffffff", "#a16207"),
}

func dark(id, name, background, surface, text, border, muted, primary, onPrimary, accent string) Theme {
	return Theme{ID: id, Name: name, Dark: true, Palette: giode.Palette{
		Background: background, Surface: surface, Text: text, Border: border,
		Muted: muted, Secondary: muted, Primary: primary, OnPrimary: onPrimary, Accent: accent,
		Success: "#4ade80", Warning: "#fbbf24", Danger: "#f87171", Info: "#60a5fa",
	}}
}

func light(id, name, background, surface, text, border, muted, primary, onPrimary, accent string) Theme {
	return Theme{ID: id, Name: name, Palette: giode.Palette{
		Background: background, Surface: surface, Text: text, Border: border,
		Muted: muted, Secondary: muted, Primary: primary, OnPrimary: onPrimary, Accent: accent,
		Success: "#16a34a", Warning: "#b45309", Danger: "#dc2626", Info: "#2563eb",
	}}
}
