package tools

import (
	"bufio"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
)

// responseText reads body and converts it to text according to the
// response's content type, which it also returns. Types that are not text
// are refused, since their bytes would only waste the model's context.
func responseText(resp *http.Response, body io.Reader) (string, string, error) {
	buffered := bufio.NewReader(body)
	mediaType := contentMediaType(resp.Header.Get("Content-Type"), buffered)
	switch {
	case mediaType == "text/html" || mediaType == "application/xhtml+xml":
		text, err := htmlToText(buffered)
		return text, mediaType, err
	case isPlainTextType(mediaType):
		data, err := io.ReadAll(buffered)
		return string(data), mediaType, err
	}
	return "", mediaType, fmt.Errorf("unsupported content type %q; web_fetch only returns HTML, JSON, XML and text", mediaType)
}

// contentMediaType parses the declared type, sniffing the body when the
// server declares none.
func contentMediaType(declared string, body *bufio.Reader) string {
	if declared == "" {
		head, _ := body.Peek(512)
		declared = http.DetectContentType(head)
	}
	mediaType, _, err := mime.ParseMediaType(declared)
	if err != nil {
		return strings.ToLower(strings.TrimSpace(declared))
	}
	return mediaType
}

// isPlainTextType reports whether a media type is readable as-is.
func isPlainTextType(mediaType string) bool {
	switch {
	case strings.HasPrefix(mediaType, "text/"),
		strings.HasSuffix(mediaType, "+json"),
		strings.HasSuffix(mediaType, "+xml"):
		return true
	}
	switch mediaType {
	case "application/json", "application/xml", "application/javascript",
		"application/x-javascript", "application/yaml", "application/x-yaml", "application/toml":
		return true
	}
	return false
}
