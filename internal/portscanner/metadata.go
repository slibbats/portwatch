package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type MetadataEntry struct {
	Port     int    `json:"port"`
	Proto    string `json:"proto"`
	Key      string `json:"key"`
	Value    string `json:"value"`
}

func metadataKey(port int, proto, key string) string {
	return fmt.Sprintf("%d/%s/%s", port, proto, key)
}

type MetadataStore struct {
	path    string
	entries map[string]MetadataEntry
}

func NewMetadataStore(dir string) (*MetadataStore, error) {
	path := filepath.Join(dir, "metadata.json")
	store := &MetadataStore{path: path, entries: make(map[string]MetadataEntry)}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return store, nil
	}
	if err != nil {
		return nil, fmt.Errorf("metadata: read: %w", err)
	}
	if err := json.Unmarshal(data, &store.entries); err != nil {
		return nil, fmt.Errorf("metadata: unmarshal: %w", err)
	}
	return store, nil
}

func (s *MetadataStore) Set(port int, proto, key, value string) error {
	s.entries[metadataKey(port, proto, key)] = MetadataEntry{Port: port, Proto: proto, Key: key, Value: value}
	return s.save()
}

func (s *MetadataStore) Get(port int, proto, key string) (string, bool) {
	e, ok := s.entries[metadataKey(port, proto, key)]
	if !ok {
		return "", false
	}
	return e.Value, true
}

func (s *MetadataStore) Remove(port int, proto, key string) error {
	k := metadataKey(port, proto, key)
	if _, ok := s.entries[k]; !ok {
		return fmt.Errorf("metadata: entry %d/%s/%s not found", port, proto, key)
	}
	delete(s.entries, k)
	return s.save()
}

func (s *MetadataStore) All() []MetadataEntry {
	out := make([]MetadataEntry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Port != out[j].Port {
			return out[i].Port < out[j].Port
		}
		return out[i].Key < out[j].Key
	})
	return out
}

func (s *MetadataStore) save() error {
	data, err := json.MarshalIndent(s.entries, "", "  ")
	if err != nil {
		return fmt.Errorf("metadata: marshal: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("metadata: mkdir: %w", err)
	}
	return os.WriteFile(s.path, data, 0o644)
}
