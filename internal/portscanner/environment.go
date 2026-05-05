package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func environmentKey(port int, proto string) string {
	return fmt.Sprintf("%d/%s", port, proto)
}

// EnvironmentEntry associates a port/protocol with a named environment (e.g. "production", "staging").
type EnvironmentEntry struct {
	Port        int    `json:"port"`
	Proto       string `json:"proto"`
	Environment string `json:"environment"`
}

// EnvironmentStore persists environment annotations for port/protocol pairs.
type EnvironmentStore struct {
	path    string
	entries map[string]EnvironmentEntry
}

// NewEnvironmentStore loads or initialises an environment store at the given directory.
func NewEnvironmentStore(dir string) (*EnvironmentStore, error) {
	s := &EnvironmentStore{
		path:    filepath.Join(dir, "environments.json"),
		entries: make(map[string]EnvironmentEntry),
	}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("environment store: load: %w", err)
	}
	return s, nil
}

// Set assigns an environment label to the given port/protocol pair.
func (s *EnvironmentStore) Set(port int, proto, environment string) error {
	s.entries[environmentKey(port, proto)] = EnvironmentEntry{
		Port:        port,
		Proto:       proto,
		Environment: environment,
	}
	return s.save()
}

// Get returns the environment entry for the given port/protocol, and whether it was found.
func (s *EnvironmentStore) Get(port int, proto string) (EnvironmentEntry, bool) {
	e, ok := s.entries[environmentKey(port, proto)]
	return e, ok
}

// Remove deletes the environment annotation for the given port/protocol pair.
func (s *EnvironmentStore) Remove(port int, proto string) error {
	delete(s.entries, environmentKey(port, proto))
	return s.save()
}

// All returns all environment entries sorted by port then protocol.
func (s *EnvironmentStore) All() []EnvironmentEntry {
	out := make([]EnvironmentEntry, 0, len(s.entries))
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

func (s *EnvironmentStore) save() error {
	data, err := json.MarshalIndent(s.entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

func (s *EnvironmentStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.entries)
}
