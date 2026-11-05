package skills

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
)

// downloadFile saves the body of a GET of rawURL to path, refusing bodies
// over the install size limit.
func downloadFile(ctx context.Context, client *http.Client, rawURL, path string) (err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close() // nothing is lost if closing a read body fails
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: %s", rawURL, resp.Status)
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	var limits budget
	return limits.copy(file, resp.Body)
}
