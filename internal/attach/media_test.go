package attach

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"
)

// wav builds a WAV of seconds of silence.
func wav(rate, channels, bits int, seconds float64) []byte {
	data := make([]byte, int(float64(rate*channels*bits/8)*seconds))
	var b bytes.Buffer
	b.WriteString("RIFF")
	binary.Write(&b, binary.LittleEndian, uint32(36+len(data)))
	b.WriteString("WAVEfmt ")
	binary.Write(&b, binary.LittleEndian, uint32(16))
	binary.Write(&b, binary.LittleEndian, uint16(1))
	binary.Write(&b, binary.LittleEndian, uint16(channels))
	binary.Write(&b, binary.LittleEndian, uint32(rate))
	binary.Write(&b, binary.LittleEndian, uint32(rate*channels*bits/8))
	binary.Write(&b, binary.LittleEndian, uint16(channels*bits/8))
	binary.Write(&b, binary.LittleEndian, uint16(bits))
	b.WriteString("data")
	binary.Write(&b, binary.LittleEndian, uint32(len(data)))
	b.Write(data)
	return b.Bytes()
}

func TestWAVAttachesWithItsDetails(t *testing.T) {
	att, err := Read("tone.wav", bytes.NewReader(wav(44100, 2, 16, 1.5)), testLimits)
	if err != nil {
		t.Fatal(err)
	}
	if att.Kind != KindDocument || att.MediaType != "audio/wav" {
		t.Fatalf("got %v %s", att.Kind, att.MediaType)
	}
	if !strings.Contains(att.Text, "44.1 kHz, 2 channel(s), 16-bit, 1.50 s") {
		t.Fatalf("details %q", att.Text)
	}
}

func TestMediaTypes(t *testing.T) {
	ftyp := func(brand string) []byte {
		return append([]byte{0, 0, 0, 24, 'f', 't', 'y', 'p'}, []byte(brand+"\x00\x00\x00\x00mp41")...)
	}
	cases := []struct {
		name string
		data []byte
		want string
	}{
		{"a.mp3", append([]byte("ID3\x04\x00\x00\x00\x00\x00\x00"), make([]byte, 64)...), "audio/mpeg"},
		{"a.mp3", append([]byte{0xFF, 0xFB, 0x90, 0x00}, make([]byte, 64)...), "audio/mpeg"},
		{"a.aac", append([]byte{0xFF, 0xF1, 0x50, 0x80}, make([]byte, 64)...), "audio/aac"},
		{"a.flac", append([]byte("fLaC"), make([]byte, 64)...), "audio/flac"},
		{"a.ogg", append([]byte("OggS\x00"), make([]byte, 64)...), "audio/ogg"},
		{"a.m4a", ftyp("M4A "), "audio/mp4"},
		{"a.mp4", ftyp("isom"), "video/mp4"},
		{"a.mov", ftyp("qt  "), "video/quicktime"},
		{"a.webm", append([]byte{0x1A, 0x45, 0xDF, 0xA3}, make([]byte, 64)...), "video/webm"},
	}
	for _, c := range cases {
		att, err := Read(c.name, bytes.NewReader(c.data), testLimits)
		if err != nil || att.MediaType != c.want || att.Kind != KindDocument {
			t.Errorf("%s: %s %v (want %s)", c.name, att.MediaType, err, c.want)
		}
	}
}
