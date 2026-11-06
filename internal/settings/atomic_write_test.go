package settings

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestWriteFileAtomicConcurrentWritersLeaveOneWholeFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "s.json")
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if err := WriteFileAtomic(path, []byte(fmt.Sprintf("writer %02d", i))); err != nil {
				t.Error(err)
			}
		}(i)
	}
	wg.Wait()
	data, err := os.ReadFile(path)
	if err != nil || len(data) != len("writer 00") {
		t.Fatalf("data=%q err=%v", data, err)
	}
	leftovers, _ := filepath.Glob(filepath.Join(filepath.Dir(path), ".tmp-*"))
	if len(leftovers) != 0 {
		t.Fatalf("temporary files left: %v", leftovers)
	}
}
