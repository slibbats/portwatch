package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type ExposureLevel string

const (
	ExposurePublic   ExposureLevel = "public"
	ExposureInternal ExposureLevel = "internal"
	ExposurePrivate  ExposureLevel = "private"
	ExposureUnknown  ExposureLevel = "unknown"
)

func ParseExposureLevel(s string) (ExposureLevel, error) {
	switch ExposureLevel(s) {
	case ExposurePublic, ExposureInternal, ExposurePrivate, ExposureUnknown:
		return ExposureLevel(s), nil
	}
	return "", fmt.Errorf("unknown exposure level %q: must be one of public, internal, private, unknown", s)
}

type ExposureEntry struct {
	Port     int           `json:"port"`
	Proto    string        `json:"proto"`
	Exposure ExposureLevel `json:"exposure"`
}

func exposureKey(port int, proto string) string {
	return fmt.Sprintf("%d/%s", port, proto)
}

type ExposureStore struct {
	dir     string
	entries map[string]ExposureEntry
}

func NewExposureStore(dir string) (*ExposureStore, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("exposure store: mkdir: %w", err)
	}
	s := &ExposureStore{dir: dir, entries: make(map[string]ExposureEntry)}
	_ = s.load()
	return s, nil
}

func (s *ExposureStore) Set(port int, proto string, level ExposureLevel) error {
	s.entries[exposureKey(port, proto)] = ExposureEntry{Port: port, Proto: proto, Exposure: level}
	return s.save()
}

func (s *ExposureStore) Get(port int, proto string) (ExposureLevel, bool) {
	e, ok := s.entries[exposureKey(port, proto)]
	if !ok {
		return ExposureUnknown, false
	}
	return e.Exposure, true
}

func (s *ExposureStore) Remove(port int, proto string) error {
	key := exposureKey(port, proto)
	if _, ok := s.entries[key]; !ok {
		return fmt.Errorf("exposure: no entry for %s", key)
	}
	delete(s.entries, key)
	return s.save()
}

func (s *ExposureStore) All() []ExposureEntry {
	out := make([]ExposureEntry, 0, len(s.entries))
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

func (s *ExposureStore) path() string {
	return filepath.Join(s.dir, "exposure.json")
}

func (s *ExposureStore) save() error {
	data, err := json.MarshalIndent(s.All(), "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path(), data, 0644)
}

func (s *ExposureStore) load() error {
	data, err := os.ReadFile(s.path())
	if err != nil {
		return err
	}
	var entries []ExposureEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return err
	}
	for _, e := range entries {
		s.entries[exposureKey(e.Port, e.Proto)] = e
	}
	return nil
}
