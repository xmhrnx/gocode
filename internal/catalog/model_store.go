package catalog

import (
	"context"
	"net/http"
	"sort"
	"sync"
	"time"

	"gocode/internal/llm"
	"gocode/internal/providers"
	"gocode/internal/secrets"
)

// ModelStatus is what is known about one provider's models.
type ModelStatus struct {
	Models  []llm.ModelInfo
	Loading bool
	Err     error
}

// ModelStore fetches and caches the model lists of providers. It is safe
// for concurrent use; onChange is called (from any goroutine) whenever a
// list changes, so the UI can redraw.
type ModelStore struct {
	keys       secrets.Store
	httpClient *http.Client
	onChange   func()

	mu         sync.Mutex
	byProvider map[string]ModelStatus
}

// NewModelStore returns an empty store.
func NewModelStore(keys secrets.Store, httpClient *http.Client, onChange func()) *ModelStore {
	return &ModelStore{keys: keys, httpClient: httpClient, onChange: onChange, byProvider: map[string]ModelStatus{}}
}

// Status returns the provider's model list as last fetched.
func (s *ModelStore) Status(providerID string) ModelStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.byProvider[providerID]
}

// Models returns the provider's models: the fetched ones, plus suggested
// and hand-added IDs the server didn't list, sorted by ID.
func (s *ModelStore) Models(def providers.Definition) []llm.ModelInfo {
	status := s.Status(def.ID)
	seen := map[string]bool{}
	models := make([]llm.ModelInfo, 0, len(status.Models)+len(def.SuggestedModels))
	for _, m := range status.Models {
		seen[m.ID] = true
		models = append(models, m)
	}
	for _, id := range def.SuggestedModels {
		if !seen[id] {
			seen[id] = true
			models = append(models, llm.ModelInfo{ID: id})
		}
	}
	sort.Slice(models, func(i, j int) bool { return models[i].ID < models[j].ID })
	return models
}

// Refresh fetches the provider's model list in the background.
func (s *ModelStore) Refresh(def providers.Definition) {
	s.update(def.ID, func(st *ModelStatus) { st.Loading = true; st.Err = nil })
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		models, err := s.fetch(ctx, def)
		s.update(def.ID, func(st *ModelStatus) {
			st.Loading, st.Err = false, err
			if err == nil {
				st.Models = models
			}
		})
	}()
}

func (s *ModelStore) fetch(ctx context.Context, def providers.Definition) ([]llm.ModelInfo, error) {
	client, err := NewClient(def, s.keys, s.httpClient)
	if err != nil {
		return nil, err
	}
	return client.ListModels(ctx)
}

func (s *ModelStore) update(providerID string, change func(*ModelStatus)) {
	s.mu.Lock()
	status := s.byProvider[providerID]
	change(&status)
	s.byProvider[providerID] = status
	s.mu.Unlock()
	if s.onChange != nil {
		s.onChange()
	}
}
