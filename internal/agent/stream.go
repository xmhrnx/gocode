package agent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gocode/internal/llm"
)

// streamReply makes one model call, retrying failures that happen before
// any output arrived, and assembles the reply message.
func streamReply(ctx context.Context, cfg Config, history []llm.Message, emit func(Event)) (llm.Message, error) {
	for attempt := 0; ; attempt++ {
		builder := newReplyBuilder()
		err := cfg.Client.Stream(ctx, cfg.request(history), func(ev llm.Event) {
			builder.add(ev, emit)
		})
		if err == nil || ctx.Err() != nil || builder.started || !retryable(err) || attempt >= cfg.MaxRetries {
			return builder.message(), err
		}
		wait := time.Duration(1<<attempt) * time.Second
		emit(Event{Kind: EventRetry, Text: fmt.Sprintf("%v. Retrying in %s…", err, wait)})
		select {
		case <-ctx.Done():
			return builder.message(), ctx.Err()
		case <-time.After(wait):
		}
	}
}

func retryable(err error) bool {
	var apiErr *llm.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Retryable()
	}
	return !errors.Is(err, context.Canceled) // network errors are worth another try
}
