package theme

import (
	"gioui.org/unit"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/styles/properties"
)

// Type holds the text sizes in use, derived from the user's base size.
// It is only read and changed on the UI goroutine.
type Type struct {
	Base, Small, Tiny, Title, Heading, Code unit.Sp
}

// Sizes is the current typography; SetBaseSize changes it.
var Sizes = typeFor(14, 13)

// SetBaseSize updates the text sizes from the UI and code font sizes.
func SetBaseSize(ui, code int) { Sizes = typeFor(unit.Sp(ui), unit.Sp(code)) }

func typeFor(base, code unit.Sp) Type {
	return Type{Base: base, Small: base - 1, Tiny: base - 2, Title: base + 2, Heading: base + 6, Code: code}
}

// lineHeight returns a comfortable line height for a font size.
func lineHeight(size unit.Sp) properties.LineHeight {
	return properties.LineHeight(size * 1.55)
}

// Body returns the style of regular text.
func Body() giode.Styles {
	return giode.Styles{Color: Text, FontSize: Sizes.Base, LineHeight: lineHeight(Sizes.Base)}
}

// Caption returns the style of small secondary text.
func Caption() giode.Styles {
	return giode.Styles{Color: Muted, FontSize: Sizes.Small}
}

// Label returns the style of a control's label.
func Label() giode.Styles {
	return giode.Styles{Color: Text, FontSize: Sizes.Base, FontWeight: giode.FontWeightMedium}
}

// Mono returns the style of code and command output.
func Mono() giode.Styles {
	return giode.Styles{Color: Text, FontSize: Sizes.Code, FontFamily: "JetBrains Mono", LineHeight: lineHeight(Sizes.Code)}
}
