package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type HostEntry struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Hostname string `json:"hostname"`
}

func hostmapKey(port int, proto string) string {
	return fmt.Sprintf("%d/%s", port, proto)
}

type HostmapStore struct {
	dir     string
	entries map[string]HostEntry
}

func NewHostmapStore(dir string) (*HostmapStore, error) {
	s := &HostmapStore{
		dir:     dir,
		entries: make(map[string]HostEntry),
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("hostmap: mkdir: %w", err)
	}
	_ = s.load()
	return s, nil
}

func (s *HostmapStore) Set(port int, proto, hostname string) error {
	key := hostmapKey(port, proto)
	s.entries[key] = HostEntry{Port: port, Protocol: proto, Hostname: hostname}
	return s.save()
}

func (s *HostmapStore) Get(port int, proto string) (string, bool) {
	e, ok := s.entries[hostmapKey(port, proto)]
	if !ok {
		return "", false
	}
	return e.Hostname, true
}

func (s *HostmapStore) Remove(port int, proto string) error {
	key := hostmapKey(port, proto)
	if _, ok := s.entries[key]; !ok {
		return fmt.Errorf("hostmap: entry %s not found", key)
	}
	delete(s.entries, key)
	return s.save()
}

func (s *HostmapStore) All() []HostEntry {
	out := make([]HostEntry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Port != out[j].Port {
			return out[i].Port < out[j].Port
		}
		return out[i].Protocol < out[j].Protocol
	})
	return out
}

func (s *HostmapStore) path() string {
	return filepath.Join(s.dir, "hostmap.json")
}

func (s *HostmapStore) save() error {
	data, err := json.MarshalIndent(s.entries, "", "  ")
	if err != nil {
		return fmt.Errorf("hostmap: marshal: %w", err)
	}
	return os.WriteFile(s.path(), data, 0644)
}

func (s *HostmapStore) load() error {
	data, err := os.ReadFile(s.path())
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.entries)
}
