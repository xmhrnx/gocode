package tools

import (
	"context"
	"io/fs"
	"path/filepath"
)

// walkFiles calls visit for every regular file under root, skipping ignored
// directories and files. Unreadable entries are skipped rather than failing
// the whole search. visit may return fs.SkipAll to stop early.
func walkFiles(ctx context.Context, root string, rules *ignoreRules, visit func(path string, d fs.DirEntry) error) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		if err != nil {
			if path == root {
				return err
			}
			return skipEntry(d)
		}
		if path != root && rules.skip(path, d.IsDir()) {
			return skipEntry(d)
		}
		if !d.Type().IsRegular() {
			return nil
		}
		return visit(path, d)
	})
}

// skipEntry skips a directory's whole subtree, or just the one file.
func skipEntry(d fs.DirEntry) error {
	if d != nil && d.IsDir() {
		return filepath.SkipDir
	}
	return nil
}
