package llm

import "context"

// Client talks to one provider.
type Client interface {
	// Stream sends the request and calls emit for each event, in order, on
	// the calling goroutine. It returns when the reply is complete.
	Stream(ctx context.Context, req Request, emit func(Event)) error
	// ListModels returns the models the provider offers.
	ListModels(ctx context.Context) ([]ModelInfo, error)
}

// ModelInfo describes one model.
type ModelInfo struct {
	ID            string `json:"id"`
	Name          string `json:"name,omitempty"`
	ContextWindow int    `json:"context_window,omitempty"`
}
