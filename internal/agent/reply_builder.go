package agent

import "gocode/internal/llm"

// replyBuilder assembles streamed events into an assistant message.
type replyBuilder struct {
	parts   llm.PartBuilder
	started bool
	// stopReason is why the model stopped, once it has.
	stopReason string
}

func newReplyBuilder() *replyBuilder { return &replyBuilder{} }

func (b *replyBuilder) add(ev llm.Event, emit func(Event)) {
	switch ev.Kind {
	case llm.EventText, llm.EventReasoning:
		b.started = true
		b.addText(ev, emit)
	case llm.EventToolCall:
		b.started = true
		b.parts.AddPart(llm.Part{Kind: llm.PartToolCall, ToolCall: ev.ToolCall})
	case llm.EventPart:
		b.started = true
		b.parts.AddPart(*ev.Part)
		emit(Event{Kind: EventPart, Part: ev.Part})
	case llm.EventStop:
		b.stopReason = ev.StopReason
	case llm.EventUsage:
		emit(Event{Kind: EventUsage, Usage: ev.Usage})
	}
}

// addText extends the current text or reasoning part, storing a provider
// signature that must be sent back unchanged with it.
func (b *replyBuilder) addText(ev llm.Event, emit func(Event)) {
	kind, kindEvent := llm.PartText, EventText
	if ev.Kind == llm.EventReasoning {
		kind, kindEvent = llm.PartReasoning, EventReasoning
	}
	if ev.Text != "" || (ev.Signature == "" && !ev.Redacted) {
		b.parts.AppendText(kind, ev.Text)
	}
	if ev.Signature != "" || ev.Redacted {
		b.parts.Sign(kind, ev.Signature, ev.Redacted)
	}
	if ev.Text != "" {
		emit(Event{Kind: kindEvent, Text: ev.Text})
	}
}

func (b *replyBuilder) message() llm.Message {
	return llm.Message{Role: llm.RoleAssistant, Parts: b.parts.Parts()}
}
