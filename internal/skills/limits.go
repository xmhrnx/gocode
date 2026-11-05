package skills

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// budget counts what an install has written so a hostile or mistaken
// source cannot fill the disk.
type budget struct {
	files int
	bytes int64
}

const (
	maxInstallFiles = 2000
	maxInstallBytes = 50 << 20
)

var errTooLarge = fmt.Errorf("skill is larger than %d MB", maxInstallBytes>>20)

// addFile counts one more file.
func (b *budget) addFile() error {
	b.files++
	if b.files > maxInstallFiles {
		return fmt.Errorf("skill has more than %d files", maxInstallFiles)
	}
	return nil
}

// copy streams src to dst, failing once the total passes the byte limit.
// The reader is capped rather than trusting declared sizes, which an
// archive can misstate.
func (b *budget) copy(dst io.Writer, src io.Reader) error {
	n, err := io.Copy(dst, io.LimitReader(src, maxInstallBytes-b.bytes+1))
	b.bytes += n
	if err != nil {
		return err
	}
	if b.bytes > maxInstallBytes {
		return errTooLarge
	}
	return nil
}

// writeFile creates target from src within the budget. Permissions are
// normalized, keeping only whether the file is executable (scripts).
func (b *budget) writeFile(target string, mode fs.FileMode, src io.Reader) (err error) {
	if err := b.addFile(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	perm := fs.FileMode(0o644)
	if mode&0o111 != 0 {
		perm = 0o755
	}
	file, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	return b.copy(file, src)
}
