package settingsview

import (
	"reflect"
	"testing"
)

func TestSplitCommandLine(t *testing.T) {
	got := splitCommandLine(`npx -y "@scope/server name" '--flag=a b' plain`)
	want := []string{"npx", "-y", "@scope/server name", "--flag=a b", "plain"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q", got)
	}
}

func TestParsePairs(t *testing.T) {
	got := parsePairs("A=1\n\nB = two=2\nnot a pair", "=")
	if !reflect.DeepEqual(got, map[string]string{"A": "1", "B": "two=2"}) {
		t.Fatalf("got %v", got)
	}
}
