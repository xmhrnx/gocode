package settingsview

import (
	"strconv"
	"strings"

	"gocode/internal/settings"
	"gocode/ui/core"
)

// providerSettings returns the user's settings for a provider, adding an
// entry on first change. The pointer is only valid until the next call,
// since adding an entry may move the slice.
func providerSettings(s *core.Services, id string) *settings.ProviderSettings {
	if ps, ok := s.Settings.Provider(id); ok {
		return ps
	}
	s.Settings.Providers = append(s.Settings.Providers, settings.ProviderSettings{ID: id})
	return &s.Settings.Providers[len(s.Settings.Providers)-1]
}

// parseModelList splits a comma or newline separated list of model IDs.
func parseModelList(text string) []string {
	var models []string
	for _, field := range strings.FieldsFunc(text, func(r rune) bool { return r == ',' || r == '\n' }) {
		if model := strings.TrimSpace(field); model != "" {
			models = append(models, model)
		}
	}
	return models
}

// customProviderID derives a unique ID from a display name.
func customProviderID(s *core.Services, name string) string {
	base := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			return r
		case r >= 'A' && r <= 'Z':
			return r + ('a' - 'A')
		}
		return '-'
	}, strings.TrimSpace(name))
	base = "custom-" + strings.Trim(base, "-")
	id := base
	for n := 2; ; n++ {
		if _, taken := s.Settings.Provider(id); !taken {
			return id
		}
		id = base + "-" + strconv.Itoa(n)
	}
}
