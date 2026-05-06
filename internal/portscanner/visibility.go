package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type Visibility string

const (
	VisibilityPublic   Visibility = "public"
	VisibilityInternal Visibility = "internal"
	VisibilityPrivate  Visibility = "private"
)

func ParseVisibility(s string) (Visibility, error) {
	switch Visibility(s) {
	case VisibilityPublic, VisibilityInternal, VisibilityPrivate:
		return Visibility(s), nil
	}
	return "", fmt.Errorf("unknown visibility %q: must be public, internal, or private", s)
}

func visibilityKey(port int, proto string) string {
	return fmt.Sprintf("%d/%s", port, proto)
}

type VisibilityStore struct {
	dir    string
	entries map[string]Visibility
}

func NewVisibilityStore(dir string) (*VisibilityStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("visibility store: mkdir: %w", err)
	}
	vs := &VisibilityStore{dir: dir, entries: make(map[string]Visibility)}
	if err := vs.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return vs, nil
}

func (vs *VisibilityStore) Set(port int, proto string, v Visibility) error {
	vs.entries[visibilityKey(port, proto)] = v
	return vs.save()
}

func (vs *VisibilityStore) Get(port int, proto string) (Visibility, bool) {
	v, ok := vs.entries[visibilityKey(port, proto)]
	return v, ok
}

func (vs *VisibilityStore) Remove(port int, proto string) error {
	key := visibilityKey(port, proto)
	if _, ok := vs.entries[key]; !ok {
		return fmt.Errorf("visibility: no entry for %s", key)
	}
	delete(vs.entries, key)
	return vs.save()
}

func (vs *VisibilityStore) All() map[string]Visibility {
	copy := make(map[string]Visibility, len(vs.entries))
	for k, v := range vs.entries {
		copy[k] = v
	}
	return copy
}

func (vs *VisibilityStore) AllSorted() []string {
	keys := make([]string, 0, len(vs.entries))
	for k := range vs.entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (vs *VisibilityStore) path() string {
	return filepath.Join(vs.dir, "visibility.json")
}

func (vs *VisibilityStore) save() error {
	data, err := json.MarshalIndent(vs.entries, "", "  ")
	if err != nil {
		return fmt.Errorf("visibility store: marshal: %w", err)
	}
	return os.WriteFile(vs.path(), data, 0o644)
}

func (vs *VisibilityStore) load() error {
	data, err := os.ReadFile(vs.path())
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &vs.entries)
}
