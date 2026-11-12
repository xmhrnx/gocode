package llm

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"mime"
	"strconv"
	"strings"
)

// Media is an image, audio clip or video a model made. It is shown in
// the chat and can be saved; only the provider that made it may get it
// back.
type Media struct {
	Name      string `json:"name"`
	MediaType string `json:"media_type"`
	Data      []byte `json:"data"`
}

// PartMedia is a Media the model made, in Part.Media.
const PartMedia PartKind = "media"

// MediaPart returns a part holding media a model made, named after its
// kind and its place in the reply (image-2.png). Raw PCM audio, which
// Gemini's speech models send, becomes a WAV file that players open.
func MediaPart(mediaType string, data []byte, index int) Part {
	if rate, ok := pcmRate(mediaType); ok {
		mediaType, data = "audio/wav", pcmToWAV(data, rate)
	}
	name := fmt.Sprintf("%s-%d%s", mediaKind(mediaType), index, mediaExtension(mediaType))
	return Part{Kind: PartMedia, Media: &Media{Name: name, MediaType: mediaType, Data: data}}
}

// IsImage reports whether the media is a picture.
func (m Media) IsImage() bool { return strings.HasPrefix(m.MediaType, "image/") }

// IsVideo reports whether the media is a video.
func (m Media) IsVideo() bool { return strings.HasPrefix(m.MediaType, "video/") }

// ParseDataURL reads a base64 data: URL, such as an image in a reply.
func ParseDataURL(url string) (mediaType string, data []byte, ok bool) {
	rest, found := strings.CutPrefix(url, "data:")
	if !found {
		return "", nil, false
	}
	header, payload, found := strings.Cut(rest, ",")
	if !found || !strings.HasSuffix(header, ";base64") {
		return "", nil, false
	}
	data, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return "", nil, false
	}
	return strings.TrimSuffix(header, ";base64"), data, true
}

// mediaKind names a file by what it holds.
func mediaKind(mediaType string) string {
	switch {
	case strings.HasPrefix(mediaType, "image/"):
		return "image"
	case strings.HasPrefix(mediaType, "audio/"):
		return "audio"
	case strings.HasPrefix(mediaType, "video/"):
		return "video"
	}
	return "file"
}

// commonExtensions are the extensions people expect, where the system's
// type table may give an older or rarer one (.jpe, .mpga).
var commonExtensions = map[string]string{
	"image/png": ".png", "image/jpeg": ".jpg", "image/webp": ".webp", "image/gif": ".gif",
	"audio/wav": ".wav", "audio/mpeg": ".mp3", "audio/mp3": ".mp3", "audio/ogg": ".ogg",
	"audio/aac": ".aac", "audio/flac": ".flac", "video/mp4": ".mp4", "video/webm": ".webm",
	"video/quicktime": ".mov",
}

func mediaExtension(mediaType string) string {
	base, _, _ := strings.Cut(mediaType, ";")
	if ext, ok := commonExtensions[base]; ok {
		return ext
	}
	if exts, err := mime.ExtensionsByType(base); err == nil && len(exts) > 0 {
		return exts[0]
	}
	return ".bin"
}

// pcmRate reports whether a type is raw 16-bit PCM (audio/L16 or
// audio/pcm) and its sample rate, 24 kHz unless it says.
func pcmRate(mediaType string) (int, bool) {
	base, params, err := mime.ParseMediaType(mediaType)
	if err != nil || (base != "audio/l16" && base != "audio/pcm") {
		return 0, false
	}
	if rate, err := strconv.Atoi(params["rate"]); err == nil && rate > 0 {
		return rate, true
	}
	return 24000, true
}

// pcmToWAV wraps mono 16-bit little-endian PCM in a WAV header.
func pcmToWAV(pcm []byte, rate int) []byte {
	out := make([]byte, 44, 44+len(pcm))
	copy(out[0:], "RIFF")
	binary.LittleEndian.PutUint32(out[4:], uint32(36+len(pcm)))
	copy(out[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(out[16:], 16)
	binary.LittleEndian.PutUint16(out[20:], 1) // PCM
	binary.LittleEndian.PutUint16(out[22:], 1) // mono
	binary.LittleEndian.PutUint32(out[24:], uint32(rate))
	binary.LittleEndian.PutUint32(out[28:], uint32(rate*2))
	binary.LittleEndian.PutUint16(out[32:], 2)
	binary.LittleEndian.PutUint16(out[34:], 16)
	copy(out[36:], "data")
	binary.LittleEndian.PutUint32(out[40:], uint32(len(pcm)))
	return append(out, pcm...)
}
