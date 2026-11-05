package llm

import "testing"

func TestPartBuilderMergesAndSeals(t *testing.T) {
	var b PartBuilder
	b.AppendText(PartReasoning, "think ")
	b.AppendText(PartReasoning, "more")
	b.Sign(PartReasoning, "sig", false)
	b.AppendText(PartText, "Hello, ")
	b.AppendText(PartText, "world")
	b.AddPart(Part{Kind: PartToolCall, ToolCall: &ToolCall{ID: "1"}})
	parts := b.Parts()
	if len(parts) != 3 {
		t.Fatalf("got %d parts", len(parts))
	}
	if parts[0].Text != "think more" || parts[0].Signature != "sig" {
		t.Errorf("reasoning part = %+v", parts[0])
	}
	if parts[1].Text != "Hello, world" {
		t.Errorf("text part = %+v", parts[1])
	}
}

func TestPartBuilderSignatureOnlyPart(t *testing.T) {
	var b PartBuilder
	b.Sign(PartReasoning, "opaque", true)
	parts := b.Parts()
	if len(parts) != 1 || !parts[0].Redacted || parts[0].Signature != "opaque" {
		t.Fatalf("parts = %+v", parts)
	}
}
