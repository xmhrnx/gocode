package attach

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	_ "image/gif" // register decoders for attached images
	"image/jpeg"
	"image/png"
	"net/http"

	_ "golang.org/x/image/bmp" // and BMP, TIFF and WebP, sent as PNG or JPEG
	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// maxImageSide is the longest side sent to models: larger images cost
// more tokens without helping, and providers downscale them anyway.
const (
	maxImageSide  = 2000
	thumbnailSide = 120
)

// isImage reports whether data is an image GoCode can read: PNG, JPEG,
// GIF, WebP, BMP or TIFF.
func isImage(data []byte) bool {
	switch http.DetectContentType(data) {
	case "image/png", "image/jpeg", "image/gif", "image/webp", "image/bmp":
		return true
	}
	return bytes.HasPrefix(data, []byte("II*\x00")) || bytes.HasPrefix(data, []byte("MM\x00*")) // TIFF
}

func readImage(name string, data []byte) (Attachment, error) {
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return Attachment{}, fmt.Errorf("decode %s: %w", name, err)
	}
	att := Attachment{Name: name, Kind: KindImage, MediaType: "image/" + format, Data: data,
		Thumbnail: scaleToFit(img, thumbnailSide)}
	if longestSide(img) > maxImageSide || (format != "png" && format != "jpeg") {
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

// FromImage makes an image attachment from a decoded image, such as one
// pasted from the clipboard, scaled down like a file would be.
func FromImage(name string, img image.Image) (Attachment, error) {
	data, mediaType, err := encode(scaleToFit(img, maxImageSide))
	if err != nil {
		return Attachment{}, fmt.Errorf("encode %s: %w", name, err)
	}
	return Attachment{Name: name, Kind: KindImage, MediaType: mediaType, Data: data,
		Thumbnail: scaleToFit(img, thumbnailSide)}, nil
}
