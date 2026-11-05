package platform

import (
	"net/http"
	"time"
)

// NewHTTPClient returns the client used for provider calls. It bounds the
// wait for a response's headers, not the whole response: a streamed reply
// legitimately lasts many minutes.
func NewHTTPClient(headerTimeout time.Duration) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = headerTimeout
	return &http.Client{Transport: transport}
}
