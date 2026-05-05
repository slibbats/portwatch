package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type categoryKey struct {
	Port  int    `json:"port"`
	Proto string `json:"proto"`
}

type CategoryStore struct {
	path       string
	categories map[categoryKey]string
}

func NewCategoryStore(dir string) (*CategoryStore, error) {
	cs := &CategoryStore{
		path:       filepath.Join(dir, "categories.json"),
		categories: make(map[categoryKey]string),
	}
	if err := cs.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("category store: load: %w", err)
	}
	return cs, nil
}

func (cs *CategoryStore) Set(port int, proto, category string) error {
	cs.categories[categoryKey{Port: port, Proto: proto}] = category
	return cs.save()
}

func (cs *CategoryStore) Get(port int, proto string) (string, bool) {
	v, ok := cs.categories[categoryKey{Port: port, Proto: proto}]
	return v, ok
}

func (cs *CategoryStore) Remove(port int, proto string) error {
	delete(cs.categories, categoryKey{Port: port, Proto: proto})
	return cs.save()
}

type categoryEntry struct {
	Port     int    `json:"port"`
	Proto    string `json:"proto"`
	Category string `json:"category"`
}

func (cs *CategoryStore) All() []categoryEntry {
	out := make([]categoryEntry, 0, len(cs.categories))
	for k, v := range cs.categories {
		out = append(out, categoryEntry{Port: k.Port, Proto: k.Proto, Category: v})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Port != out[j].Port {
			return out[i].Port < out[j].Port
		}
		return out[i].Proto < out[j].Proto
	})
	return out
}

func (cs *CategoryStore) save() error {
	data, err := json.MarshalIndent(cs.All(), "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cs.path, data, 0644)
}

func (cs *CategoryStore) load() error {
	data, err := os.ReadFile(cs.path)
	if err != nil {
		return err
	}
	var entries []categoryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return err
	}
	for _, e := range entries {
		cs.categories[categoryKey{Port: e.Port, Proto: e.Proto}] = e.Category
	}
	return nil
}
