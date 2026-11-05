package theme

import "github.com/xmhrnx/giode"

// Semantic colors. Mixed from theme tokens, so they follow theme switches.
var (
	Canvas      = "theme.background"
	Panel       = "theme.surface"
	Text        = "theme.text"
	Muted       = "theme.muted"
	Border      = "theme.border"
	Primary     = "theme.primary"
	OnPrimary   = "theme.onprimary"
	Danger      = "theme.danger"
	Success     = "theme.success"
	Warning     = "theme.warning"
	Faint       = giode.Mix("theme.muted", "theme.surface", 0.7)
	Raised      = giode.Mix("theme.text", "theme.surface", 0.04)
	Hover       = giode.Mix("theme.text", "theme.surface", 0.07)
	Pressed     = giode.Mix("theme.text", "theme.surface", 0.11)
	SoftBorder  = giode.Mix("theme.border", "theme.surface", 0.6)
	PrimarySoft = giode.Mix("theme.primary", "theme.surface", 0.14)
	DangerSoft  = giode.Mix("theme.danger", "theme.surface", 0.12)
	WarningSoft = giode.Mix("theme.warning", "theme.surface", 0.12)
	SuccessSoft = giode.Mix("theme.success", "theme.surface", 0.12)
	CanvasHover = giode.Mix("theme.text", "theme.background", 0.05)
	CodeBack    = giode.Mix("theme.text", "theme.background", 0.035)
)
