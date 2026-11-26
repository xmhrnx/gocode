package remoteui

import (
	"encoding/json"
	"errors"
	"slices"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/settings"

	"gocode/internal/remote/client"
	"gocode/internal/remote/secure"
)

// identityName is where the phone's key pair is kept in secure storage.
const identityName = "phone.identity"

// desktopsKey is where the paired desktops are kept in the settings.
const desktopsKey = "desktops"

// store keeps the phone's identity and the desktops it is paired with.
// Its methods run on the UI goroutine, except loadIdentity.
type store struct {
	key      secure.KeyPair
	desktops []client.Desktop
}

// loadIdentity reads the phone's key pair, making it the first time. It
// can block on the platform's key store: call it off the UI goroutine.
func loadIdentity() (secure.KeyPair, error) {
	var key secure.KeyPair
	raw, err := giode.SecureStore.Get(identityName)
	if err == nil && json.Unmarshal(raw, &key) == nil && key.Valid() {
		return key, nil
	}
	if err != nil && !errors.Is(err, giode.ErrNotFound) {
		return key, err
	}
	if key, err = secure.NewKeyPair(); err != nil {
		return key, err
	}
	data, _ := json.Marshal(key)
	return key, giode.SecureStore.Set(identityName, data)
}

func loadDesktops() []client.Desktop {
	return settings.Value(desktopsKey, []client.Desktop{})
}

// save writes the desktops.
func (s *store) save() { _ = settings.Save(desktopsKey, s.desktops) }

// put adds or replaces the desktop with the same key.
func (s *store) put(d client.Desktop) {
	s.desktops = slices.DeleteFunc(s.desktops, func(x client.Desktop) bool { return x.Key == d.Key })
	s.desktops = append(s.desktops, d)
	s.save()
}

// remove forgets a desktop.
func (s *store) remove(key string) {
	s.desktops = slices.DeleteFunc(s.desktops, func(x client.Desktop) bool { return x.Key == key })
	s.save()
}

// find returns the desktop with key, or nil.
func (s *store) find(key string) *client.Desktop {
	for i := range s.desktops {
		if s.desktops[i].Key == key {
			return &s.desktops[i]
		}
	}
	return nil
}

// phoneName is what desktops call this phone.
func phoneName() string {
	if model := giode.Device().Model; model != "" {
		return model
	}
	return "Phone"
}
