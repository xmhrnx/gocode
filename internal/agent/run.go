package agent

import (
	"context"
	"fmt"

	"gocode/internal/llm"
)

// runner is one run of the agent loop.
type runner struct {
	cfg     Config
	history []llm.Message
	added   []llm.Message
	emit    func(Event)
	guard   loopGuard
	// lastInputTokens is the size of the last request, which tells when
	// the conversation nears the context window.
	lastInputTokens int
}

// Run continues the conversation in history until the model replies
// without calling tools, the step limit (if any) is reached, or ctx ends.
// It reports progress through emit, on the calling goroutine, and returns
// the messages it added. On cancellation the messages added so far are
// returned with ctx's error, and every tool call left has a result, so
// the conversation stays valid for the next request.
func Run(ctx context.Context, cfg Config, history []llm.Message, emit func(Event)) ([]llm.Message, error) {
	r := &runner{cfg: cfg, history: history, emit: emit}
	err := r.loop(ctx)
	return r.added, err
}

func (r *runner) loop(ctx context.Context) error {
	for step := 0; r.cfg.MaxSteps <= 0 || step < r.cfg.MaxSteps; step++ {
		if r.cfg.needsCompaction(r.lastInputTokens) {
			if err := r.compact(ctx); err != nil {
				return err
			}
		}
		reply, err := r.callModel(ctx)
		if len(reply.Parts) > 0 {
			r.append(reply)
		}
		calls := reply.ToolCalls()
		if err != nil {
			if len(calls) > 0 {
				r.append(cancelledResults(calls))
			}
			return err
		}
		if len(calls) == 0 {
			return nil
		}
		results, err := runToolCalls(ctx, r.cfg, calls, &r.guard, r.emit)
		r.append(results)
		if err != nil {
			return err
		}
	}
	r.emit(Event{Kind: EventNotice, Text: fmt.Sprintf("Stopped after %d steps (the limit set in Settings → Agent). Send a message to continue.", r.cfg.MaxSteps)})
	return nil
}

// callModel streams one reply. A request rejected as too long for the
// context window is compacted and sent once more.
func (r *runner) callModel(ctx context.Context) (llm.Message, error) {
	reply, err := streamReply(ctx, r.cfg, r.history, r.track)
	if err == nil || len(reply.Parts) > 0 || r.cfg.CompactAtPercent <= 0 || !isContextOverflow(err) {
		return reply, err
	}
	if compactErr := r.compact(ctx); compactErr != nil {
		return reply, fmt.Errorf("%w (compacting to make it fit failed too: %v)", err, compactErr)
	}
	return streamReply(ctx, r.cfg, r.history, r.track)
}

func (r *runner) compact(ctx context.Context) error {
	history, summary, err := compact(ctx, r.cfg, r.history)
	if err != nil {
		return err
	}
	r.history, r.lastInputTokens = history, 0
	r.emit(Event{Kind: EventCompacted, Text: summary})
	return nil
}

// track passes events on, remembering the request size from usage.
func (r *runner) track(ev Event) {
	if ev.Kind == EventUsage && ev.Usage != nil {
		r.lastInputTokens = ev.Usage.InputTokens + ev.Usage.CacheReadTokens + ev.Usage.CacheWriteTokens
	}
	r.emit(ev)
}

func (r *runner) append(msg llm.Message) {
	r.history = append(r.history, msg)
	r.added = append(r.added, msg)
	r.emit(Event{Kind: EventMessage, Message: &msg})
}
