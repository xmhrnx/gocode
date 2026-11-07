package websearch

import (
	"context"
	"sync"
)

// lazy sets up its searcher on the first search, so reading the key from
// the keychain happens when a search runs, not while a run is prepared.
type lazy struct {
	once     sync.Once
	setup    func() (Searcher, error)
	searcher Searcher
	err      error
}

// Lazy returns a searcher that calls setup on its first search.
func Lazy(setup func() (Searcher, error)) Searcher { return &lazy{setup: setup} }

func (l *lazy) Search(ctx context.Context, query string, count int) ([]Result, error) {
	l.once.Do(func() { l.searcher, l.err = l.setup() })
	if l.err != nil {
		return nil, l.err
	}
	return l.searcher.Search(ctx, query, count)
}
