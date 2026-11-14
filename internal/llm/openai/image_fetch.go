package openai

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
)

// maxFetchedImage is the largest image a reply may link to that is
// downloaded into the chat.
const maxFetchedImage = 40 << 20

// imageFetcher downloads an image a reply links to instead of sending.
type imageFetcher func(url string) (mediaType string, data []byte, err error)

// fetchImage downloads an image a provider returned as a link, which may
// expire, so the chat keeps it like one sent inline.
func (c *Client) fetchImage(ctx context.Context) imageFetcher {
	return func(url string) (string, []byte, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return "", nil, err
		}
		resp, err := c.http.Do(req)
		if err != nil {
			return "", nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", nil, fmt.Errorf("%s", resp.Status)
		}
		mediaType, _, _ := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if !strings.HasPrefix(mediaType, "image/") {
			return "", nil, fmt.Errorf("not an image: %q", mediaType)
		}
		data, err := io.ReadAll(io.LimitReader(resp.Body, maxFetchedImage+1))
		if err == nil && len(data) > maxFetchedImage {
			err = fmt.Errorf("image larger than %d MB", maxFetchedImage>>20)
		}
		return mediaType, data, err
	}
}
