// Package assets embeds the fonts and icons GoCode ships with.
package assets

import (
	"embed"
	"fmt"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/styles/properties"
)

// fontFile names an embedded font and the family and weight it provides.
type fontFile struct {
	path   string
	family string
	weight properties.FontWeight
}

//go:embed fonts/*.ttf icons/*.svg
var files embed.FS

// AppIcon is the window and taskbar icon, a PNG.
//
//go:embed icon.png
var AppIcon []byte

// Font families registered by Load.
const (
	UIFont   = "Inter"
	CodeFont = "JetBrains Mono"
)

var fontFiles = []fontFile{
	{"fonts/Inter-Regular.ttf", UIFont, giode.FontWeightNormal},
	{"fonts/Inter-Medium.ttf", UIFont, giode.FontWeightMedium},
	{"fonts/Inter-SemiBold.ttf", UIFont, giode.FontWeightSemiBold},
	{"fonts/Inter-Bold.ttf", UIFont, giode.FontWeightBold},
	{"fonts/JetBrainsMono-Regular.ttf", CodeFont, giode.FontWeightNormal},
	{"fonts/JetBrainsMono-Bold.ttf", CodeFont, giode.FontWeightBold},
}

// Load registers the fonts and icons with Giode. Icons are named after
// their file, e.g. "copy" for icons/copy.svg.
func Load() error {
	fonts := make([]giode.FontFile, 0, len(fontFiles))
	for _, f := range fontFiles {
		data, err := files.ReadFile(f.path)
		if err != nil {
			return fmt.Errorf("read %s: %w", f.path, err)
		}
		fonts = append(fonts, giode.FontFile{Family: f.family, Weight: f.weight, Data: data})
	}
	if err := giode.RegisterFont(fonts...); err != nil {
		return fmt.Errorf("load fonts: %w", err)
	}
	if err := giode.RegisterIconsFS(files, "icons"); err != nil {
		return fmt.Errorf("load icons: %w", err)
	}
	return nil
}
