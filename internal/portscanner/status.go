package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type PortStatus string

const (
	StatusActive   PortStatus = "active"
	StatusInactive PortStatus = "inactive"
	StatusUnknown  PortStatus = "unknown"
)

type StatusEntry struct {
	Port      int        `json:"port"`
	Proto     string     `json:"proto"`
	Status    PortStatus `json:"status"`
	UpdatedAt time.Time  `json:"updated_at"`
	Note      string     `json:"note,omitempty"`
}

type StatusStore struct {
	path    string
	entries map[string]StatusEntry
}

func statusKey(port int, proto string) string {
	return fmt.Sprintf("%d/%s", port, proto)
}

func NewStatusStore(dir string) (*StatusStore, error) {
	s := &StatusStore{
		path:    filepath.Join(dir, "status.json"),
		entries: make(map[string]StatusEntry),
	}
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, fmt.Errorf("status: read: %w", err)
	}
	if err := json.Unmarshal(data, &s.entries); err != nil {
		return nil, fmt.Errorf("status: unmarshal: %w", err)
	}
	return s, nil
}

func (s *StatusStore) Set(port int, proto string, status PortStatus, note string) error {
	key := statusKey(port, proto)
	s.entries[key] = StatusEntry{
		Port:      port,
		Proto:     proto,
		Status:    status,
		UpdatedAt: time.Now().UTC(),
		Note:      note,
	}
	return s.save()
}

func (s *StatusStore) Get(port int, proto string) (StatusEntry, bool) {
	e, ok := s.entries[statusKey(port, proto)]
	return e, ok
}

func (s *StatusStore) Remove(port int, proto string) error {
	delete(s.entries, statusKey(port, proto))
	return s.save()
}

func (s *StatusStore) All() []StatusEntry {
	out := make([]StatusEntry, 0, len(s.entries))
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

func (s *StatusStore) save() error {
	data, err := json.MarshalIndent(s.entries, "", "  ")
	if err != nil {
		return fmt.Errorf("status: marshal: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("status: mkdir: %w", err)
	}
	return os.WriteFile(s.path, data, 0o644)
}
