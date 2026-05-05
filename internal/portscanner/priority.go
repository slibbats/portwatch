package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type Priority int

const (
	PriorityLow    Priority = 1
	PriorityMedium Priority = 2
	PriorityHigh   Priority = 3
	PriorityCritical Priority = 4
)

func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "low"
	case PriorityMedium:
		return "medium"
	case PriorityHigh:
		return "high"
	case PriorityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

func ParsePriority(s string) (Priority, error) {
	switch s {
	case "low":
		return PriorityLow, nil
	case "medium":
		return PriorityMedium, nil
	case "high":
		return PriorityHigh, nil
	case "critical":
		return PriorityCritical, nil
	default:
		return 0, fmt.Errorf("unknown priority %q: must be low, medium, high, or critical", s)
	}
}

type PriorityEntry struct {
	Port     int      `json:"port"`
	Proto    string   `json:"proto"`
	Priority Priority `json:"priority"`
}

func priorityKey(port int, proto string) string {
	return fmt.Sprintf("%d/%s", port, proto)
}

type PriorityStore struct {
	path    string
	entries map[string]PriorityEntry
}

func NewPriorityStore(dir string) (*PriorityStore, error) {
	ps := &PriorityStore{
		path:    filepath.Join(dir, "priorities.json"),
		entries: make(map[string]PriorityEntry),
	}
	data, err := os.ReadFile(ps.path)
	if err != nil {
		if os.IsNotExist(err) {
			return ps, nil
		}
		return nil, fmt.Errorf("read priority store: %w", err)
	}
	if err := json.Unmarshal(data, &ps.entries); err != nil {
		return nil, fmt.Errorf("parse priority store: %w", err)
	}
	return ps, nil
}

func (ps *PriorityStore) Set(port int, proto string, p Priority) error {
	ps.entries[priorityKey(port, proto)] = PriorityEntry{Port: port, Proto: proto, Priority: p}
	return ps.save()
}

func (ps *PriorityStore) Get(port int, proto string) (Priority, bool) {
	e, ok := ps.entries[priorityKey(port, proto)]
	if !ok {
		return 0, false
	}
	return e.Priority, true
}

func (ps *PriorityStore) Remove(port int, proto string) error {
	key := priorityKey(port, proto)
	if _, ok := ps.entries[key]; !ok {
		return fmt.Errorf("no priority set for %s", key)
	}
	delete(ps.entries, key)
	return ps.save()
}

func (ps *PriorityStore) All() []PriorityEntry {
	out := make([]PriorityEntry, 0, len(ps.entries))
	for _, e := range ps.entries {
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

func (ps *PriorityStore) save() error {
	if err := os.MkdirAll(filepath.Dir(ps.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(ps.entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ps.path, data, 0o644)
}
