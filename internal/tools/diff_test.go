package tools

import (
	"math/rand"
	"strings"
	"testing"
)

// rebuild returns the old and new texts an edit script describes.
func rebuild(ops []lineOp) (oldLines, newLines []string) {
	for _, op := range ops {
		if op.kind != opInsert {
			oldLines = append(oldLines, op.text)
		}
		if op.kind != opDelete {
			newLines = append(newLines, op.text)
		}
	}
	return oldLines, newLines
}

func countChanges(ops []lineOp) int {
	n := 0
	for _, op := range ops {
		if op.kind != opEqual {
			n++
		}
	}
	return n
}

func TestDiffLinesMinimal(t *testing.T) {
	tests := []struct {
		a, b    string
		changes int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "a b", 2},
		{"a b c", "a b c", 0},
		{"a b c a b b a", "c b a b a c", 5},
		{"a x c", "a y c", 2},
		{"a b c d", "b c d e", 2},
	}
	for _, tt := range tests {
		a, b := strings.Fields(tt.a), strings.Fields(tt.b)
		ops := diffLines(a, b)
		gotA, gotB := rebuild(ops)
		if strings.Join(gotA, " ") != tt.a || strings.Join(gotB, " ") != tt.b {
			t.Errorf("diff(%q,%q) rebuilds %v / %v", tt.a, tt.b, gotA, gotB)
		}
		if got := countChanges(ops); got != tt.changes {
			t.Errorf("diff(%q,%q) has %d changes, want %d", tt.a, tt.b, got, tt.changes)
		}
	}
}

func TestDiffLinesRandomRoundTrip(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	randomLines := func() []string {
		lines := make([]string, rng.Intn(30))
		for i := range lines {
			lines[i] = string(rune('a' + rng.Intn(4)))
		}
		return lines
	}
	for i := 0; i < 500; i++ {
		a, b := randomLines(), randomLines()
		gotA, gotB := rebuild(diffLines(a, b))
		if strings.Join(gotA, "") != strings.Join(a, "") || strings.Join(gotB, "") != strings.Join(b, "") {
			t.Fatalf("round trip failed for %v -> %v", a, b)
		}
	}
}

func TestUnifiedDiff(t *testing.T) {
	old := "1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12\n"
	updated := "1\n2\nthree\n4\n5\n6\n7\n8\n9\n10\n11\n12\n13"
	want := `--- a/f
+++ b/f
@@ -1,6 +1,6 @@
 1
 2
-3
+three
 4
 5
 6
@@ -10,3 +10,4 @@
 10
 11
 12
+13
\ No newline at end of file
`
	if got := unifiedDiff("a/f", "b/f", old, updated); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestUnifiedDiffMergesCloseChanges(t *testing.T) {
	got := unifiedDiff("a", "b", "1\n2\n3\n4\n5\n6\n7\n8\n", "x\n2\n3\n4\n5\n6\n7\ny\n")
	if strings.Count(got, "@@ -") != 1 || !strings.Contains(got, "@@ -1,8 +1,8 @@") {
		t.Errorf("expected one merged hunk, got:\n%s", got)
	}
}

func TestUnifiedDiffNewFileAndEqual(t *testing.T) {
	if got := unifiedDiff("a", "b", "same\n", "same\n"); got != "" {
		t.Errorf("equal texts should give no diff, got %q", got)
	}
	got := unifiedDiff("/dev/null", "b/new", "", "hello\n")
	if !strings.Contains(got, "@@ -0,0 +1,1 @@\n+hello\n") {
		t.Errorf("unexpected new-file diff:\n%s", got)
	}
}

func TestDiffLinesFallsBackWhenTooDifferent(t *testing.T) {
	var a, b []string
	for i := 0; i < maxEditDistance; i++ {
		a = append(a, "old"+strings.Repeat("x", i%7))
		b = append(b, "new"+strings.Repeat("y", i%5))
	}
	gotA, gotB := rebuild(diffLines(a, b))
	if len(gotA) != len(a) || len(gotB) != len(b) || gotA[0] != a[0] || gotB[len(b)-1] != b[len(b)-1] {
		t.Errorf("fallback diff does not rebuild the inputs")
	}
}
