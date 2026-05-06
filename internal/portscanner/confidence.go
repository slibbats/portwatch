package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type ConfidenceLevel int

const (
	ConfidenceLow    ConfidenceLevel = 1
	ConfidenceMedium ConfidenceLevel = 2
	ConfidenceHigh   ConfidenceLevel = 3
)

func (c ConfidenceLevel) String() string {
	switch c {
	case ConfidenceLow:
		return "low"
	case ConfidenceMedium:
		return "medium"
	case ConfidenceHigh:
		return "high"
	default:
		return "unknown"
	}
}

func ParseConfidence(s string) (ConfidenceLevel, error) {
	switch s {
	case "low":
		return ConfidenceLow, nil
	case "medium":
		return ConfidenceMedium, nil
	case "high":
		return ConfidenceHigh, nil
	default:
		return 0, fmt.Errorf("unknown confidence level: %q (want low, medium, high)", s)
	}
}

type ConfidenceEntry struct {
	Port     int             `json:"port"`
	Proto    string          `json:"proto"`
	Level    ConfidenceLevel `json:"level"`
	Rationale string        `json:"rationale,omitempty"`
}

func confidenceKey(port int, proto string) string {
	return fmt.Sprintf("%d/%s", port, proto)
}

type ConfidenceStore struct {
	dir     string
	entries map[string]ConfidenceEntry
}

func NewConfidenceStore(dir string) (*ConfidenceStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("confidence store: mkdir: %w", err)
	}
	cs := &ConfidenceStore{dir: dir, entries: make(map[string]ConfidenceEntry)}
	_ = cs.load()
	return cs, nil
}

func (cs *ConfidenceStore) Set(port int, proto string, level ConfidenceLevel, rationale string) error {
	cs.entries[confidenceKey(port, proto)] = ConfidenceEntry{
		Port: port, Proto: proto, Level: level, Rationale: rationale,
	}
	return cs.save()
}

func (cs *ConfidenceStore) Get(port int, proto string) (ConfidenceEntry, bool) {
	e, ok := cs.entries[confidenceKey(port, proto)]
	return e, ok
}

func (cs *ConfidenceStore) Remove(port int, proto string) error {
	key := confidenceKey(port, proto)
	if _, ok := cs.entries[key]; !ok {
		return fmt.Errorf("confidence: no entry for %s", key)
	}
	delete(cs.entries, key)
	return cs.save()
}

func (cs *ConfidenceStore) All() []ConfidenceEntry {
	out := make([]ConfidenceEntry, 0, len(cs.entries))
	for _, e := range cs.entries {
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

func (cs *ConfidenceStore) save() error {
	path := filepath.Join(cs.dir, "confidence.json")
	data, err := json.MarshalIndent(cs.entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (cs *ConfidenceStore) load() error {
	path := filepath.Join(cs.dir, "confidence.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &cs.entries)
}
