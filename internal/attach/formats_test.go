package attach

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"strings"
	"testing"

	"golang.org/x/image/bmp"
)

var testLimits = Limits{MaxBytes: 20 << 20, MaxTextBytes: 256 << 10}

// zipped builds a zip archive of files.
func zipped(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, content := range files {
		f, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		f.Write([]byte(content))
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// minimalPDF builds a one-page PDF showing text, with a correct xref.
func minimalPDF(text string) []byte {
	stream := fmt.Sprintf("BT /F1 12 Tf 72 712 Td (%s) Tj ET", text)
	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(stream), stream),
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
	}
	var b bytes.Buffer
	b.WriteString("%PDF-1.4\n")
	offsets := make([]int, len(objects))
	for i, obj := range objects {
		offsets[i] = b.Len()
		fmt.Fprintf(&b, "%d 0 obj\n%s\nendobj\n", i+1, obj)
	}
	xref := b.Len()
	fmt.Fprintf(&b, "xref\n0 %d\n0000000000 65535 f \n", len(objects)+1)
	for _, off := range offsets {
		fmt.Fprintf(&b, "%010d 00000 n \n", off)
	}
	fmt.Fprintf(&b, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xref)
	return b.Bytes()
}

func TestPDFKeepsTheFileAndItsText(t *testing.T) {
	att, err := Read("report.pdf", bytes.NewReader(minimalPDF("Quarterly results")), testLimits)
	if err != nil {
		t.Fatal(err)
	}
	if att.Kind != KindDocument || att.MediaType != "application/pdf" || !bytes.HasPrefix(att.Data, []byte("%PDF")) {
		t.Fatalf("attachment %+v", att.Kind)
	}
	if !strings.Contains(strings.ReplaceAll(att.Text, " ", ""), "Quarterlyresults") {
		t.Fatalf("text %q", att.Text)
	}
	msg := BuildMessage("summarize", []Attachment{att})
	if msg.Parts[0].Document == nil || msg.Parts[0].Document.Name != "report.pdf" {
		t.Fatalf("message %+v", msg.Parts)
	}
}

func TestOfficeAndBookFilesBecomeText(t *testing.T) {
	cases := map[string]map[string]string{
		"notes.docx": {"word/document.xml": `<w:document xmlns:w="w"><w:body><w:p><w:r><w:t>Hello</w:t></w:r><w:r><w:tab/><w:t>world</w:t></w:r></w:p><w:p><w:r><w:t>Second</w:t></w:r></w:p></w:body></w:document>`},
		"deck.pptx": {
			"ppt/slides/slide2.xml":  `<p:sld xmlns:a="a" xmlns:p="p"><a:p><a:r><a:t>Second</a:t></a:r></a:p></p:sld>`,
			"ppt/slides/slide1.xml":  `<p:sld xmlns:a="a" xmlns:p="p"><a:p><a:r><a:t>Hello world</a:t></a:r></a:p></p:sld>`,
			"ppt/slides/slide10.xml": `<p:sld xmlns:a="a" xmlns:p="p"><a:p><a:r><a:t>Tenth</a:t></a:r></a:p></p:sld>`,
		},
		"sheet.xlsx": {
			"xl/sharedStrings.xml":     `<sst><si><t>Hello</t></si><si><r><t>wor</t></r><r><t>ld</t></r></si></sst>`,
			"xl/worksheets/sheet1.xml": `<worksheet><sheetData><row><c t="s"><v>0</v></c><c t="s"><v>1</v></c><c><v>42</v></c></row><row><c t="inlineStr"><is><t>Second</t></is></c></row></sheetData></worksheet>`,
		},
		"doc.odt": {"content.xml": `<office:document-content xmlns:office="o" xmlns:text="t"><office:body><office:text><text:p>Hello<text:tab/>world</text:p><text:p>Second</text:p></office:text></office:body></office:document-content>`},
		"book.epub": {
			"META-INF/container.xml": `<container><rootfiles><rootfile full-path="OEBPS/content.opf"/></rootfiles></container>`,
			"OEBPS/content.opf":      `<package><manifest><item id="c1" href="one.xhtml"/><item id="c2" href="two.xhtml"/></manifest><spine><itemref idref="c1"/><itemref idref="c2"/></spine></package>`,
			"OEBPS/one.xhtml":        `<html><head><title>skip me</title></head><body><p>Hello world</p></body></html>`,
			"OEBPS/two.xhtml":        `<html><body><p>Second</p><script>skip()</script></body></html>`,
		},
	}
	for name, files := range cases {
		att, err := Read(name, bytes.NewReader(zipped(t, files)), testLimits)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		text := strings.Join(strings.Fields(string(att.Data)), " ")
		if att.Kind != KindText || !strings.Contains(text, "Hello") || !strings.Contains(text, "world") || !strings.Contains(text, "Second") {
			t.Errorf("%s: %q", name, text)
		}
		if strings.Contains(text, "skip") {
			t.Errorf("%s kept text it should skip: %q", name, text)
		}
	}
}

func TestSlidesAreInOrder(t *testing.T) {
	data := zipped(t, map[string]string{
		"ppt/slides/slide2.xml":  `<p:sld xmlns:a="a" xmlns:p="p"><a:t>two</a:t></p:sld>`,
		"ppt/slides/slide10.xml": `<p:sld xmlns:a="a" xmlns:p="p"><a:t>ten</a:t></p:sld>`,
		"ppt/slides/slide1.xml":  `<p:sld xmlns:a="a" xmlns:p="p"><a:t>one</a:t></p:sld>`,
	})
	att, err := Read("deck.pptx", bytes.NewReader(data), testLimits)
	if err != nil {
		t.Fatal(err)
	}
	text := string(att.Data)
	if !(strings.Index(text, "one") < strings.Index(text, "two") && strings.Index(text, "two") < strings.Index(text, "ten")) {
		t.Fatalf("order: %q", text)
	}
}

func TestUTF16TextAndBMPImages(t *testing.T) {
	utf16 := []byte{0xFF, 0xFE, 'H', 0, 'i', 0, ' ', 0, 0xE9, 0}
	att, err := Read("notes.txt", bytes.NewReader(utf16), testLimits)
	if err != nil || string(att.Data) != "Hi é" {
		t.Fatalf("utf-16: %q %v", att.Data, err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	img.Set(1, 1, color.RGBA{255, 0, 0, 255})
	var buf bytes.Buffer
	if err := bmp.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	att, err = Read("pic.bmp", &buf, testLimits)
	if err != nil || att.Kind != KindImage || (att.MediaType != "image/png" && att.MediaType != "image/jpeg") {
		t.Fatalf("bmp: %v %v %s", err, att.Kind, att.MediaType)
	}
}

func TestUnsupportedFilesSaySo(t *testing.T) {
	_, err := Read("tool.exe", bytes.NewReader([]byte{0x4D, 0x5A, 0x90, 0, 3, 0, 0, 0}), testLimits)
	if err == nil || !strings.Contains(err.Error(), "PDF") {
		t.Fatalf("error %v", err)
	}
}
