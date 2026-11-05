// Package jsonstr decodes JSON string literals without repairing invalid
// UTF-8.
//
// Some OpenAI-compatible servers stream tokens that end in the middle of a
// multi-byte character, so one fragment holds the first bytes of "→" and
// the next fragment the rest. encoding/json turns each half into U+FFFD,
// which corrupts the text for good. Decoding with the raw bytes kept lets
// the halves join back into the character once concatenated.
package jsonstr

import (
	"strconv"
	"unicode/utf16"
	"unicode/utf8"
)

// simpleEscapes maps the letter after a backslash to the byte it stands for.
var simpleEscapes = map[byte]byte{'"': '"', '\\': '\\', '/': '/', 'b': '\b', 'f': '\f', 'n': '\n', 'r': '\r', 't': '\t'}

// Decode returns the content of the JSON string literal raw, keeping any
// bytes that aren't valid UTF-8. ok is false if raw isn't a string literal.
func Decode(raw []byte) (string, bool) {
	if len(raw) < 2 || raw[0] != '"' || raw[len(raw)-1] != '"' {
		return "", false
	}
	body := raw[1 : len(raw)-1]
	out := make([]byte, 0, len(body))
	for i := 0; i < len(body); i++ {
		if body[i] != '\\' {
			out = append(out, body[i])
			continue
		}
		if i+1 >= len(body) {
			return "", false
		}
		i++
		var ok bool
		if out, i, ok = appendEscape(out, body, i); !ok {
			return "", false
		}
	}
	return string(out), true
}

// appendEscape decodes the escape whose letter is at body[i] and returns
// the index of its last byte.
func appendEscape(out, body []byte, i int) ([]byte, int, bool) {
	if c, ok := simpleEscapes[body[i]]; ok {
		return append(out, c), i, true
	}
	if body[i] != 'u' {
		return out, i, false
	}
	r, ok := hex4(body, i+1)
	if !ok {
		return out, i, false
	}
	i += 4
	if utf16.IsSurrogate(r) {
		if low, ok := lowSurrogate(body, i+1); ok {
			r = utf16.DecodeRune(r, low)
			i += 6
		} else {
			r = utf8.RuneError
		}
	}
	return utf8.AppendRune(out, r), i, true
}

func lowSurrogate(body []byte, i int) (rune, bool) {
	if i+1 >= len(body) || body[i] != '\\' || body[i+1] != 'u' {
		return 0, false
	}
	return hex4(body, i+2)
}

func hex4(body []byte, i int) (rune, bool) {
	if i+4 > len(body) {
		return 0, false
	}
	n, err := strconv.ParseUint(string(body[i:i+4]), 16, 32)
	return rune(n), err == nil
}
