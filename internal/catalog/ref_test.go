package catalog

import "testing"

func TestParseRefKeepsSlashesInModel(t *testing.T) {
	ref := ParseRef("openrouter/anthropic/claude-opus-5-5")
	if ref.Provider != "openrouter" || ref.Model != "anthropic/claude-opus-5-5" {
		t.Fatalf("ref = %+v", ref)
	}
	if ref.String() != "openrouter/anthropic/claude-opus-5-5" {
		t.Fatalf("string = %s", ref.String())
	}
	if !ParseRef("nomodel").IsZero() {
		t.Fatal("a ref without a slash should be zero")
	}
}
