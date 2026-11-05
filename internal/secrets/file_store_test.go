package secrets

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestFileStoreRoundTrip(t *testing.T) {
	store := &FileStore{Path: filepath.Join(t.TempDir(), "keys.json")}
	if _, err := store.Get("a"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
	if err := store.Set("a", "secret"); err != nil {
		t.Fatal(err)
	}
	if got, _ := store.Get("a"); got != "secret" {
		t.Fatalf("got %q", got)
	}
	if err := store.Delete("a"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get("a"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("not deleted: %v", err)
	}
}
