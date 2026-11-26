package phone

import (
	"encoding/json"
	"os"

	"gocode/internal/remote/pair"
	"gocode/internal/remote/secure"
	"gocode/internal/secrets"
)

// identityKey is where the desktop's identity is kept in the key store.
const identityKey = "phone.identity"

// identity is the desktop's key pair and relay room, made once.
type identity struct {
	Key  secure.KeyPair `json:"key"`
	Room string         `json:"room"`
}

// LoadIdentity returns the desktop's identity, making and saving it the
// first time.
func LoadIdentity(keys secrets.Store) (secure.KeyPair, string, error) {
	var id identity
	if raw, err := keys.Get(identityKey); err == nil && raw != "" {
		if json.Unmarshal([]byte(raw), &id) == nil && id.Key.Valid() && id.Room != "" {
			return id.Key, id.Room, nil
		}
	}
	key, err := secure.NewKeyPair()
	if err != nil {
		return secure.KeyPair{}, "", err
	}
	id = identity{Key: key, Room: pair.NewSecret()}
	data, _ := json.Marshal(id)
	return key, id.Room, keys.Set(identityKey, string(data))
}

// DesktopName is what phones call this computer.
func DesktopName() string {
	if name, err := os.Hostname(); err == nil && name != "" {
		return name
	}
	return "Desktop"
}
