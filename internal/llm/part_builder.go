package llm

import "strings"

// PartBuilder assembles streamed chunks into message parts. Consecutive
// chunks of the same kind grow one part through a strings.Builder, so a
// long reply costs linear time and memory instead of copying the whole
// text on every chunk.
type PartBuilder struct {
	parts []Part
	// open accumulates the text of the last part while it can still grow.
	open *strings.Builder
}

// AppendText adds a text or reasoning chunk.
func (b *PartBuilder) AppendText(kind PartKind, text string) {
	if n := len(b.parts); n == 0 || b.open == nil || b.parts[n-1].Kind != kind {
		b.seal()
		b.parts = append(b.parts, Part{Kind: kind})
		b.open = &strings.Builder{}
	}
	b.open.WriteString(text)
}

// Sign stores a provider signature on the last part of kind, which closes
// it: a signed part must be sent back exactly as it was. A redacted block,
// or a signature with no unsigned part of its kind to go on, gets a part
// of its own.
func (b *PartBuilder) Sign(kind PartKind, signature string, redacted bool) {
	n := len(b.parts)
	if n == 0 || redacted || b.parts[n-1].Kind != kind || b.parts[n-1].Signature != "" {
		b.seal()
		b.parts = append(b.parts, Part{Kind: kind})
	}
	b.seal()
	last := &b.parts[len(b.parts)-1]
	last.Signature, last.Redacted = signature, redacted
}

// AddPart appends a complete part, such as a tool call.
func (b *PartBuilder) AddPart(part Part) {
	b.seal()
	b.parts = append(b.parts, part)
}

// Parts returns the parts so far. The text of the open part is shared
// with the builder, not copied.
func (b *PartBuilder) Parts() []Part {
	parts := make([]Part, len(b.parts))
	copy(parts, b.parts)
	if b.open != nil {
		parts[len(parts)-1].Text = b.open.String()
	}
	return parts
}

// Len returns the number of parts.
func (b *PartBuilder) Len() int { return len(b.parts) }

// Part returns part i, with its current text.
func (b *PartBuilder) Part(i int) Part {
	part := b.parts[i]
	if i == len(b.parts)-1 && b.open != nil {
		part.Text = b.open.String()
	}
	return part
}

// seal fixes the open part's text so a new part can start.
func (b *PartBuilder) seal() {
	if b.open != nil {
		b.parts[len(b.parts)-1].Text = b.open.String()
		b.open = nil
	}
}
