// Package fsretry retries file operations that Windows refuses for a
// moment while another program has the file open: Defender and the
// search indexer briefly lock files that were just written, and a rename
// onto a file someone is reading fails until they're done.
package fsretry

import "time"

// attempts and pause bound the retrying to about a second in all.
const (
	attempts = 8
	pause    = 25 * time.Millisecond
)

// Do runs op, retrying with growing pauses while it fails with an error
// that may clear by itself.
func Do(op func() error) error {
	err := op()
	for attempt := 1; err != nil && transient(err) && attempt < attempts; attempt++ {
		time.Sleep(time.Duration(attempt) * pause)
		err = op()
	}
	return err
}
