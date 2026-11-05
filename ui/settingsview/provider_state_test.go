package settingsview

import (
	"reflect"
	"testing"
)

func TestParseModelList(t *testing.T) {
	got := parseModelList(" a, b\n\nc ,")
	if want := []string{"a", "b", "c"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v", got)
	}
}
