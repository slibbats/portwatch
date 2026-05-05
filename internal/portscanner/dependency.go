package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type Dependency struct {
	Port     int    `json:"port"`
	Proto    string `json:"proto"`
	DependsOn []DependencyRef `json:"depends_on"`
}

type DependencyRef struct {
	Port  int    `json:"port"`
	Proto string `json:"proto"`
	Note  string `json:"note,omitempty"`
}

func dependencyKey(port int, proto string) string {
	return fmt.Sprintf("%d/%s", port, proto)
}

type DependencyStore struct {
	path  string
	entries map[string]Dependency
}

func NewDependencyStore(dir string) (*DependencyStore, error) {
	path := filepath.Join(dir, "dependencies.json")
	ds := &DependencyStore{path: path, entries: make(map[string]Dependency)}
	if err := ds.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return ds, nil
}

func (ds *DependencyStore) Set(port int, proto string, deps []DependencyRef) error {
	key := dependencyKey(port, proto)
	ds.entries[key] = Dependency{Port: port, Proto: proto, DependsOn: deps}
	return ds.save()
}

func (ds *DependencyStore) Get(port int, proto string) (Dependency, bool) {
	d, ok := ds.entries[dependencyKey(port, proto)]
	return d, ok
}

func (ds *DependencyStore) Remove(port int, proto string) error {
	delete(ds.entries, dependencyKey(port, proto))
	return ds.save()
}

func (ds *DependencyStore) All() []Dependency {
	out := make([]Dependency, 0, len(ds.entries))
	for _, d := range ds.entries {
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Port != out[j].Port {
			return out[i].Port < out[j].Port
		}
		return out[i].Proto < out[j].Proto
	})
	return out
}

func (ds *DependencyStore) save() error {
	data, err := json.MarshalIndent(ds.entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ds.path, data, 0644)
}

func (ds *DependencyStore) load() error {
	data, err := os.ReadFile(ds.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &ds.entries)
}
