package catalog

import (
	"gocode/internal/providers"
	"gocode/internal/settings"
)

// Providers returns every provider the user can configure: the built-in
// ones with the user's overrides applied, then the user's custom ones.
func Providers(s *settings.Settings) []providers.Definition {
	var defs []providers.Definition
	for _, def := range providers.Catalog() {
		if ps, ok := s.Provider(def.ID); ok {
			def = applyOverrides(def, *ps)
		}
		defs = append(defs, def)
	}
	for _, ps := range s.Providers {
		if ps.Custom {
			defs = append(defs, customDefinition(ps))
		}
	}
	return defs
}

// Find returns the provider with the given ID.
func Find(s *settings.Settings, id string) (providers.Definition, bool) {
	for _, def := range Providers(s) {
		if def.ID == id {
			return def, true
		}
	}
	return providers.Definition{}, false
}

// Enabled reports whether the user turned the provider on.
func Enabled(s *settings.Settings, id string) bool {
	ps, ok := s.Provider(id)
	return ok && ps.Enabled
}

func applyOverrides(def providers.Definition, ps settings.ProviderSettings) providers.Definition {
	if ps.BaseURL != "" {
		def.BaseURL = ps.BaseURL
	}
	if len(ps.Headers) > 0 {
		merged := map[string]string{}
		for k, v := range def.Headers {
			merged[k] = v
		}
		for k, v := range ps.Headers {
			merged[k] = v
		}
		def.Headers = merged
	}
	def.SuggestedModels = append(append([]string{}, def.SuggestedModels...), ps.ExtraModels...)
	return def
}

func customDefinition(ps settings.ProviderSettings) providers.Definition {
	style := providers.Style(ps.Style)
	if style == "" {
		style = providers.StyleOpenAI
	}
	return providers.Definition{
		ID: ps.ID, Name: ps.Name, Style: style, BaseURL: ps.BaseURL, Category: "custom",
		KeyOptional: true, Headers: ps.Headers, MaxTokensField: "max_tokens",
		SuggestedModels: ps.ExtraModels,
	}
}
