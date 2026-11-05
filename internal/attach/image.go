package attach

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	_ "image/gif" // register decoders for attached images
	"image/jpeg"
	"image/png"

	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

// maxImageSide is the longest side sent to models: larger images cost
// more tokens without helping, and providers downscale them anyway.
const (
	maxImageSide  = 2000
	thumbnailSide = 120
)

func readImage(name string, data []byte) (Attachment, error) {
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return Attachment{}, fmt.Errorf("decode %s: %w", name, err)
	}
	att := Attachment{Name: name, Kind: KindImage, MediaType: "image/" + format, Data: data,
		Thumbnail: scaleToFit(img, thumbnailSide)}
	if longestSide(img) > maxImageSide || format == "webp" || format == "gif" {
		att.Data, att.MediaType, err = encode(scaleToFit(img, maxImageSide))
		if err != nil {
			return Attachment{}, fmt.Errorf("encode %s: %w", name, err)
		}
	}
	return att, nil
}

func encode(img image.Image) ([]byte, string, error) {
	var buf bytes.Buffer
	if opaque, ok := img.(interface{ Opaque() bool }); ok && opaque.Opaque() {
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 88})
		return buf.Bytes(), "image/jpeg", err
	}
	err := png.Encode(&buf, img)
	return buf.Bytes(), "image/png", err
}

func scaleToFit(img image.Image, side int) image.Image {
	bounds := img.Bounds()
	longest := longestSide(img)
	if longest <= side {
		return img
	}
	w, h := bounds.Dx()*side/longest, bounds.Dy()*side/longest
	dst := image.NewRGBA(image.Rect(0, 0, max(w, 1), max(h, 1)))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Src, nil)
	return dst
}

func longestSide(img image.Image) int {
	return max(img.Bounds().Dx(), img.Bounds().Dy())
}

// Preview decodes image data and scales it to fit side, for display.
func Preview(data []byte, side int) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return scaleToFit(img, side), nil
}
