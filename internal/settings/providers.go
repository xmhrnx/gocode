package settings

// ProviderSettings is the user's configuration of one provider. Built-in
// providers are listed by ID with overrides; custom ones carry everything.
// API keys are not stored here but in the system keychain.
type ProviderSettings struct {
	ID      string            `json:"id"`
	Name    string            `json:"name,omitempty"`
	Style   string            `json:"style,omitempty"`
	BaseURL string            `json:"base_url,omitempty"`
	Enabled bool              `json:"enabled"`
	Custom  bool              `json:"custom,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	// ExtraModels are model IDs added by hand, for servers that don't
	// list their models.
	ExtraModels []string `json:"extra_models,omitempty"`
}

// Models holds model selection preferences.
type Models struct {
	// Default is the model new chats start with, as "provider/model".
	Default string `json:"default"`
	// TitleModel generates chat titles; empty uses the chat's model.
	TitleModel string `json:"title_model"`
	// Recent lists recently used models, newest first.
	Recent []string `json:"recent"`
	// Hidden lists models left out of the picker.
	Hidden []string `json:"hidden"`
}

// Provider returns the settings for the provider with the given ID.
func (s *Settings) Provider(id string) (*ProviderSettings, bool) {
	for i := range s.Providers {
		if s.Providers[i].ID == id {
			return &s.Providers[i], true
		}
	}
	return nil, false
}
