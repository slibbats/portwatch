package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type ImpactLevel string

const (
	ImpactCritical ImpactLevel = "critical"
	ImpactHigh     ImpactLevel = "high"
	ImpactMedium   ImpactLevel = "medium"
	ImpactLow      ImpactLevel = "low"
	ImpactNone     ImpactLevel = "none"
)

func ParseImpactLevel(s string) (ImpactLevel, error) {
	switch ImpactLevel(s) {
	case ImpactCritical, ImpactHigh, ImpactMedium, ImpactLow, ImpactNone:
		return ImpactLevel(s), nil
	}
	return "", fmt.Errorf("unknown impact level: %q (want critical|high|medium|low|none)", s)
}

type ImpactEntry struct {
	Port     int         `json:"port"`
	Proto    string      `json:"proto"`
	Level    ImpactLevel `json:"level"`
	Rationale string    `json:"rationale,omitempty"`
}

func impactKey(port int, proto string) string {
	return fmt.Sprintf("%d/%s", port, proto)
}

type ImpactStore struct {
	dir  string
	data map[string]ImpactEntry
}

func NewImpactStore(dir string) (*ImpactStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("impact store: mkdir: %w", err)
	}
	s := &ImpactStore{dir: dir, data: make(map[string]ImpactEntry)}
	_ = s.load()
	return s, nil
}

func (s *ImpactStore) Set(port int, proto string, level ImpactLevel, rationale string) error {
	s.data[impactKey(port, proto)] = ImpactEntry{Port: port, Proto: proto, Level: level, Rationale: rationale}
	return s.save()
}

func (s *ImpactStore) Get(port int, proto string) (ImpactEntry, bool) {
	e, ok := s.data[impactKey(port, proto)]
	return e, ok
}

func (s *ImpactStore) Remove(port int, proto string) error {
	key := impactKey(port, proto)
	if _, ok := s.data[key]; !ok {
		return fmt.Errorf("impact: no entry for %s", key)
	}
	delete(s.data, key)
	return s.save()
}

func (s *ImpactStore) All() []ImpactEntry {
	out := make([]ImpactEntry, 0, len(s.data))
	for _, e := range s.data {
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

func (s *ImpactStore) save() error {
	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.dir, "impact.json"), b, 0o644)
}

func (s *ImpactStore) load() error {
	b, err := os.ReadFile(filepath.Join(s.dir, "impact.json"))
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &s.data)
}
