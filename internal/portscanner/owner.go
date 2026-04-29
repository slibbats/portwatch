package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

const ownerFile = "owners.json"

func ownerKey(port int, proto string) string {
	return fmt.Sprintf("%d/%s", port, proto)
}

// OwnerEntry holds ownership metadata for a port.
type OwnerEntry struct {
	Port  int    `json:"port"`
	Proto string `json:"proto"`
	Owner string `json:"owner"`
	Team  string `json:"team,omitempty"`
	Email string `json:"email,omitempty"`
}

// OwnerStore manages port ownership records.
type OwnerStore struct {
	mu      sync.RWMutex
	entries map[string]OwnerEntry
	dir     string
}

// NewOwnerStore creates a new OwnerStore backed by dir.
func NewOwnerStore(dir string) (*OwnerStore, error) {
	s := &OwnerStore{dir: dir, entries: make(map[string]OwnerEntry)}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return s, nil
}

// Set adds or updates an ownership entry.
func (s *OwnerStore) Set(e OwnerEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[ownerKey(e.Port, e.Proto)] = e
	return s.save()
}

// Get returns the ownership entry for the given port/proto.
func (s *OwnerStore) Get(port int, proto string) (OwnerEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[ownerKey(port, proto)]
	return e, ok
}

// Remove deletes the ownership entry for the given port/proto.
func (s *OwnerStore) Remove(port int, proto string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, ownerKey(port, proto))
	return s.save()
}

// All returns all ownership entries sorted by port.
func (s *OwnerStore) All() []OwnerEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]OwnerEntry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Port < out[j].Port })
	return out
}

func (s *OwnerStore) save() error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.dir, ownerFile), data, 0o644)
}

func (s *OwnerStore) load() error {
	data, err := os.ReadFile(filepath.Join(s.dir, ownerFile))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.entries)
}
