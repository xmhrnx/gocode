// Package assets embeds the fonts and icons GoCode ships with.
package assets

import (
	"embed"
	"fmt"

	"gioui.org/font"
	"gioui.org/font/opentype"
	"github.com/xmhrnx/giode"
)

type fontFile struct {
	path   string
	family string
	weight font.Weight
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
	{"fonts/Inter-Regular.ttf", UIFont, font.Normal},
	{"fonts/Inter-Medium.ttf", UIFont, font.Medium},
	{"fonts/Inter-SemiBold.ttf", UIFont, font.SemiBold},
	{"fonts/Inter-Bold.ttf", UIFont, font.Bold},
	{"fonts/JetBrainsMono-Regular.ttf", CodeFont, font.Normal},
	{"fonts/JetBrainsMono-Bold.ttf", CodeFont, font.Bold},
}

// Load registers the fonts and icons with Giode. Icons are named after
// their file, e.g. "copy" for icons/copy.svg.
func Load() error {
	var faces []font.FontFace
	for _, f := range fontFiles {
		face, err := loadFace(f)
		if err != nil {
			return err
		}
		faces = append(faces, face)
	}
	giode.RegisterFonts(faces)
	if err := giode.RegisterIconsFS(files, "icons"); err != nil {
		return fmt.Errorf("load icons: %w", err)
	}
	return nil
}

func loadFace(f fontFile) (font.FontFace, error) {
	data, err := files.ReadFile(f.path)
	if err != nil {
		return font.FontFace{}, fmt.Errorf("read %s: %w", f.path, err)
	}
	face, err := opentype.Parse(data)
	if err != nil {
		return font.FontFace{}, fmt.Errorf("parse %s: %w", f.path, err)
	}
	return font.FontFace{
		Font: font.Font{Typeface: font.Typeface(f.family), Weight: f.weight},
		Face: face,
	}, nil
}
