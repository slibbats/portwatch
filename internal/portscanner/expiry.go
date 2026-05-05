package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type ExpiryEntry struct {
	Port     int       `json:"port"`
	Proto    string    `json:"proto"`
	ExpiresAt time.Time `json:"expires_at"`
	Note     string    `json:"note,omitempty"`
}

func expiryKey(port int, proto string) string {
	return fmt.Sprintf("%d/%s", port, proto)
}

type ExpiryStore struct {
	path    string
	entries map[string]ExpiryEntry
}

func NewExpiryStore(dir string) (*ExpiryStore, error) {
	s := &ExpiryStore{
		path:    filepath.Join(dir, "expiry.json"),
		entries: make(map[string]ExpiryEntry),
	}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return s, nil
}

func (s *ExpiryStore) Set(port int, proto string, expiresAt time.Time, note string) error {
	s.entries[expiryKey(port, proto)] = ExpiryEntry{
		Port:      port,
		Proto:     proto,
		ExpiresAt: expiresAt,
		Note:      note,
	}
	return s.save()
}

func (s *ExpiryStore) Get(port int, proto string) (ExpiryEntry, bool) {
	e, ok := s.entries[expiryKey(port, proto)]
	return e, ok
}

func (s *ExpiryStore) Remove(port int, proto string) error {
	key := expiryKey(port, proto)
	if _, ok := s.entries[key]; !ok {
		return fmt.Errorf("no expiry entry for %s", key)
	}
	delete(s.entries, key)
	return s.save()
}

func (s *ExpiryStore) Expired() []ExpiryEntry {
	now := time.Now()
	var out []ExpiryEntry
	for _, e := range s.entries {
		if now.After(e.ExpiresAt) {
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Port != out[j].Port {
			return out[i].Port < out[j].Port
		}
		return out[i].Proto < out[j].Proto
	})
	return out
}

func (s *ExpiryStore) All() []ExpiryEntry {
	out := make([]ExpiryEntry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Port != out[j].Port {
			return out[i].Port < out[j].Port
		}
		return out[i].Proto < out[j].Proto
	})
	return out
}

// PurgeExpired removes all expired entries from the store and saves it.
// It returns the list of entries that were removed.
func (s *ExpiryStore) PurgeExpired() ([]ExpiryEntry, error) {
	expired := s.Expired()
	for _, e := range expired {
		delete(s.entries, expiryKey(e.Port, e.Proto))
	}
	if len(expired) > 0 {
		if err := s.save(); err != nil {
			return nil, fmt.Errorf("purge expired: %w", err)
		}
	}
	return expired, nil
}

func (s *ExpiryStore) save() error {
	data, err := json.MarshalIndent(s.entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

func (s *ExpiryStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.entries)
}
