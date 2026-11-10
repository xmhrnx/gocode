package attach

import (
	"bytes"
	"encoding/binary"
	"unicode/utf16"
	"unicode/utf8"
)

// decodeText returns data as UTF-8 text: UTF-8 as is (without a byte
// order mark), UTF-16 with a byte order mark converted. ok is false for
// data that isn't text.
func decodeText(data []byte) (text []byte, ok bool) {
	switch {
	case bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}):
		data = data[3:]
	case bytes.HasPrefix(data, []byte{0xFF, 0xFE}):
		return fromUTF16(data[2:], binary.LittleEndian), true
	case bytes.HasPrefix(data, []byte{0xFE, 0xFF}):
		return fromUTF16(data[2:], binary.BigEndian), true
	}
	sample := data[:min(len(data), 8192)]
	if bytes.Contains(sample, []byte{0}) || !utf8.Valid(trimPartialRune(sample)) {
		return nil, false
	}
	return data, true
}

func fromUTF16(data []byte, order binary.ByteOrder) []byte {
	units := make([]uint16, len(data)/2)
	for i := range units {
		units[i] = order.Uint16(data[2*i:])
	}
	return []byte(string(utf16.Decode(units)))
}

// trimPartialRune drops a rune cut in half by sampling.
func trimPartialRune(b []byte) []byte {
	for i := 0; i < utf8.UTFMax && len(b) > 0 && !utf8.Valid(b); i++ {
		b = b[:len(b)-1]
	}
	return b
}
