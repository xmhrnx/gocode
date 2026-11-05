package settingsview

import (
	"slices"

	"gocode/internal/secrets"
	"gocode/internal/settings"
)

// saveKey stores a provider's key, enables the provider, and tests it.
func (p *providersPage) saveKey(id, key string) {
	s := p.services
	go func() {
		err := s.Keys.Set(secrets.ProviderKeyName(id), key)
		s.App.Post(func() {
			if err != nil {
				s.Error("Couldn't save the key: " + err.Error())
				return
			}
			p.hasKey[id] = true
			providerSettings(s, id).Enabled = true
			s.SettingsChanged()
			if row, ok := p.rows[id]; ok {
				row.enabled.SetChecked(true)
			}
			p.test(id)
		})
	}()
}

// addCustom adds a provider the user defined and opens its editor.
func (p *providersPage) addCustom(ps settings.ProviderSettings, key string) {
	s := p.services
	ps.ID = customProviderID(s, ps.Name)
	ps.Custom, ps.Enabled = true, true
	s.Settings.Providers = append(s.Settings.Providers, ps)
	s.SettingsChanged()
	if key != "" {
		p.saveKey(ps.ID, key)
	} else {
		p.test(ps.ID)
	}
}

// removeCustom deletes a custom provider and its key.
func (p *providersPage) removeCustom(id string) {
	s := p.services
	s.Settings.Providers = slices.DeleteFunc(s.Settings.Providers, func(ps settings.ProviderSettings) bool {
		return ps.ID == id && ps.Custom
	})
	s.SettingsChanged()
	go func() {
		if err := s.Keys.Delete(secrets.ProviderKeyName(id)); err != nil {
			s.Error("Couldn't delete the key: " + err.Error())
		}
	}()
}
