package theme

import "github.com/xmhrnx/giode"

// MenuStyle returns the style of popup menus.
func MenuStyle() giode.Styles {
	return giode.Styles{Background: Panel, Color: Text, BorderRadius: 10, FontSize: Sizes.Small}
}
