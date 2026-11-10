package attach

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net/http"
	"path"
	"strings"
)

// mediaType reports the audio or video type of data, by its content
// (and, where a container holds either, the file's extension).
func mediaType(name string, data []byte) (string, bool) {
	ext := strings.ToLower(path.Ext(name))
	switch sniffed := http.DetectContentType(data); {
	case sniffed == "audio/wave":
		return "audio/wav", true
	case sniffed == "audio/mpeg", sniffed == "audio/aiff", sniffed == "audio/basic", sniffed == "audio/midi":
		return sniffed, true
	case sniffed == "application/ogg":
		return "audio/ogg", true
	case sniffed == "video/avi":
		return "video/avi", true
	case sniffed == "video/webm" && ext == ".mkv":
		return "video/x-matroska", true
	case sniffed == "video/webm":
		return "video/webm", true
	}
	switch {
	case bytes.HasPrefix(data, []byte("fLaC")):
		return "audio/flac", true
	case frameExts[ext] && len(data) > 2 && data[0] == 0xFF && data[1]&0xE0 == 0xE0: // a bare MPEG audio frame
		if data[1]&0x06 == 0 {
			return "audio/aac", true // ADTS
		}
		return "audio/mpeg", true
	case len(data) > 12 && string(data[4:8]) == "ftyp":
		return isoMediaType(string(data[8:12])), true
	case bytes.HasPrefix(data, []byte("FLV")):
		return "video/x-flv", true
	case bytes.HasPrefix(data, []byte{0, 0, 1, 0xBA}):
		return "video/mpeg", true
	case len(data) > 188 && data[0] == 0x47 && data[188] == 0x47:
		return "video/mp2t", true
	case bytes.HasPrefix(data, []byte{0x30, 0x26, 0xB2, 0x75}): // ASF
		if ext == ".wma" {
			return "audio/x-ms-wma", true
		}
		return "video/x-ms-wmv", true
	}
	return "", false
}

// frameExts are the extensions of audio files that may start with a bare
// MPEG frame rather than a header; the frame's bytes alone could be
// other data, such as UTF-16 text.
var frameExts = map[string]bool{".mp3": true, ".mp2": true, ".aac": true}

// isoMediaType reads an MP4-family file's brand.
func isoMediaType(brand string) string {
	switch {
	case brand == "M4A " || brand == "M4B ":
		return "audio/mp4"
	case brand == "qt  ":
		return "video/quicktime"
	case strings.HasPrefix(brand, "3g"):
		return "video/3gpp"
	}
	return "video/mp4"
}

// readMedia keeps an audio or video file whole, described in words for
// models that can't take it.
func readMedia(name, mediaType string, data []byte) Attachment {
	kind := "Audio"
	if strings.HasPrefix(mediaType, "video/") {
		kind = "Video"
	}
	text := fmt.Sprintf("%s file, %s, %s.", kind, mediaType, sizeText(len(data)))
	if details, ok := wavDetails(data); ok {
		text = fmt.Sprintf("Audio file, WAV: %s, %s.", details, sizeText(len(data)))
	}
	return Attachment{Name: name, Kind: KindDocument, MediaType: mediaType, Data: data, Text: text}
}

// wavDetails describes a WAV's format and length.
func wavDetails(data []byte) (string, bool) {
	if len(data) < 12 || string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return "", false
	}
	var channels, bits, rate, dataBytes int
	for p := 12; p+8 <= len(data); {
		id, n := string(data[p:p+4]), int(binary.LittleEndian.Uint32(data[p+4:p+8]))
		body := data[p+8:]
		switch {
		case id == "fmt " && len(body) >= 16:
			channels = int(binary.LittleEndian.Uint16(body[2:4]))
			rate = int(binary.LittleEndian.Uint32(body[4:8]))
			bits = int(binary.LittleEndian.Uint16(body[14:16]))
		case id == "data":
			dataBytes = min(n, len(body))
		}
		p += 8 + n + n&1
	}
	if rate == 0 || channels == 0 || bits == 0 {
		return "", false
	}
	seconds := float64(dataBytes) / float64(rate*channels*bits/8)
	return fmt.Sprintf("%.1f kHz, %d channel(s), %d-bit, %.2f s", float64(rate)/1000, channels, bits, seconds), true
}

func sizeText(n int) string {
	if n >= 1<<20 {
		return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
	}
	return fmt.Sprintf("%d KB", max(1, n>>10))
}
