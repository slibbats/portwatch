package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type ClassificationLevel string

const (
	ClassificationPublic       ClassificationLevel = "public"
	ClassificationInternal     ClassificationLevel = "internal"
	ClassificationConfidential ClassificationLevel = "confidential"
	ClassificationRestricted   ClassificationLevel = "restricted"
)

func ParseClassificationLevel(s string) (ClassificationLevel, error) {
	switch ClassificationLevel(s) {
	case ClassificationPublic, ClassificationInternal, ClassificationConfidential, ClassificationRestricted:
		return ClassificationLevel(s), nil
	}
	return "", fmt.Errorf("unknown classification level: %q", s)
}

type ClassificationEntry struct {
	Port     int                 `json:"port"`
	Proto    string              `json:"proto"`
	Level    ClassificationLevel `json:"level"`
	Rationale string             `json:"rationale,omitempty"`
}

func classificationKey(port int, proto string) string {
	return fmt.Sprintf("%d/%s", port, proto)
}

type ClassificationStore struct {
	dir     string
	entries map[string]ClassificationEntry
}

func NewClassificationStore(dir string) (*ClassificationStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("classification store: mkdir: %w", err)
	}
	cs := &ClassificationStore{dir: dir, entries: make(map[string]ClassificationEntry)}
	_ = cs.load()
	return cs, nil
}

func (cs *ClassificationStore) Set(port int, proto string, level ClassificationLevel, rationale string) error {
	cs.entries[classificationKey(port, proto)] = ClassificationEntry{
		Port: port, Proto: proto, Level: level, Rationale: rationale,
	}
	return cs.save()
}

func (cs *ClassificationStore) Get(port int, proto string) (ClassificationEntry, bool) {
	e, ok := cs.entries[classificationKey(port, proto)]
	return e, ok
}

func (cs *ClassificationStore) Remove(port int, proto string) error {
	key := classificationKey(port, proto)
	if _, ok := cs.entries[key]; !ok {
		return fmt.Errorf("classification: no entry for %s", key)
	}
	delete(cs.entries, key)
	return cs.save()
}

func (cs *ClassificationStore) All() []ClassificationEntry {
	out := make([]ClassificationEntry, 0, len(cs.entries))
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

func (cs *ClassificationStore) file() string {
	return filepath.Join(cs.dir, "classification.json")
}

func (cs *ClassificationStore) save() error {
	data, err := json.MarshalIndent(cs.entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cs.file(), data, 0o644)
}

func (cs *ClassificationStore) load() error {
	data, err := os.ReadFile(cs.file())
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &cs.entries)
}
