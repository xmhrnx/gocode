package attach

import (
	"bytes"
	"image"
	"image/png"
	"strings"
	"testing"
)

var limits = Limits{MaxBytes: 1 << 20, MaxTextBytes: 1 << 10}

func TestReadText(t *testing.T) {
	att, err := Read("a.go", strings.NewReader("package a\n"), limits)
	if err != nil || att.Kind != KindText {
		t.Fatalf("att=%+v err=%v", att, err)
	}
}

func TestReadRejectsBinary(t *testing.T) {
	if _, err := Read("a.bin", bytes.NewReader([]byte{1, 0, 2, 3}), limits); err == nil {
		t.Fatal("want an error for binary data")
	}
}

func TestReadImageScalesLargeImages(t *testing.T) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 4000, 1000))); err != nil {
		t.Fatal(err)
	}
	att, err := Read("big.png", &buf, Limits{MaxBytes: 10 << 20, MaxTextBytes: 1})
	if err != nil {
		t.Fatal(err)
	}
	decoded, _, err := image.Decode(bytes.NewReader(att.Data))
	if err != nil || decoded.Bounds().Dx() != maxImageSide {
		t.Fatalf("width = %d, err=%v", decoded.Bounds().Dx(), err)
	}
}

func TestFileBlockRoundTrip(t *testing.T) {
	msg := BuildMessage("explain", []Attachment{{Name: "x.txt", Kind: KindText, Data: []byte("hi")}})
	name, ok := FileBlockName(msg.Parts[0].Text)
	if !ok || name != "x.txt" || msg.Parts[1].Text != "explain" {
		t.Fatalf("parts = %+v", msg.Parts)
	}
}

func TestFromImageEncodesAndScales(t *testing.T) {
	att, err := FromImage("pasted.png", image.NewRGBA(image.Rect(0, 0, 3000, 100)))
	if err != nil || att.Kind != KindImage || att.Thumbnail.Bounds().Dx() != thumbnailSide {
		t.Fatalf("att=%+v err=%v", att, err)
	}
	decoded, _, err := image.Decode(bytes.NewReader(att.Data))
	if err != nil || decoded.Bounds().Dx() != maxImageSide {
		t.Fatalf("decoded width %v, err %v", decoded.Bounds(), err)
	}
}
