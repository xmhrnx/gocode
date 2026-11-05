package theme

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/styles/properties"
)

// Icon returns a registered icon at the given size and color; an empty
// color inherits the surrounding text color.
func Icon(name string, size int, color string) giode.Element {
	return giode.Icon(name, giode.Styles{Width: properties.Width(size), Height: properties.Height(size), Color: color})
}
