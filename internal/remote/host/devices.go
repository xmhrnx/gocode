package host

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gocode/internal/remote/secure"
)

// Device is a paired phone.
type Device struct {
	Name     string    `json:"name"`
	Key      string    `json:"key"`
	Paired   time.Time `json:"paired"`
	LastSeen time.Time `json:"last_seen,omitempty"`
}

// Devices is the list of paired phones, kept in a file. It is safe for
// concurrent use.
type Devices struct {
	path string
	mu   sync.Mutex
	list []Device
}

// OpenDevices reads the list at path; a missing file is an empty list.
func OpenDevices(path string) (*Devices, error) {
	d := &Devices{path: path}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return d, nil
	}
	if err != nil {
		return d, err
	}
	return d, json.Unmarshal(data, &d.list)
}

// List returns the paired phones.
func (d *Devices) List() []Device {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]Device(nil), d.list...)
}

// paired reports whether key belongs to a paired phone, and notes that
// it was seen.
func (d *Devices) paired(key []byte) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i, dev := range d.list {
		if k, err := secure.DecodeKey(dev.Key); err == nil && bytes.Equal(k, key) {
			d.list[i].LastSeen = time.Now()
			return true
		}
	}
	return false
}

// add pairs a phone, replacing one with the same key.
func (d *Devices) add(name string, key []byte) error {
	d.mu.Lock()
	encoded := secure.EncodeKey(key)
	kept := d.list[:0]
	for _, dev := range d.list {
		if dev.Key != encoded {
			kept = append(kept, dev)
		}
	}
	now := time.Now()
	d.list = append(kept, Device{Name: name, Key: encoded, Paired: now, LastSeen: now})
	d.mu.Unlock()
	return d.save()
}

// Remove unpairs the phone with key.
func (d *Devices) Remove(key string) error {
	d.mu.Lock()
	kept := d.list[:0]
	for _, dev := range d.list {
		if dev.Key != key {
			kept = append(kept, dev)
		}
	}
	d.list = kept
	d.mu.Unlock()
	return d.save()
}

func (d *Devices) save() error {
	d.mu.Lock()
	data, err := json.MarshalIndent(d.list, "", "  ")
	d.mu.Unlock()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(d.path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(d.path, data, 0o600)
}
