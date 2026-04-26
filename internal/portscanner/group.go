package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type Group struct {
	Name  string `json:"name"`
	Ports []int  `json:"ports"`
}

type GroupStore struct {
	path   string
	groups map[string]*Group
}

func groupKey(name string) string {
	return name
}

func NewGroupStore(path string) (*GroupStore, error) {
	gs := &GroupStore{
		path:   path,
		groups: make(map[string]*Group),
	}
	if err := gs.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("groupstore: load: %w", err)
	}
	return gs, nil
}

func (gs *GroupStore) Add(name string, port int) {
	key := groupKey(name)
	g, ok := gs.groups[key]
	if !ok {
		g = &Group{Name: name}
		gs.groups[key] = g
	}
	for _, p := range g.Ports {
		if p == port {
			return
		}
	}
	g.Ports = append(g.Ports, port)
	sort.Ints(g.Ports)
}

func (gs *GroupStore) Remove(name string, port int) {
	key := groupKey(name)
	g, ok := gs.groups[key]
	if !ok {
		return
	}
	filtered := g.Ports[:0]
	for _, p := range g.Ports {
		if p != port {
			filtered = append(filtered, p)
		}
	}
	g.Ports = filtered
	if len(g.Ports) == 0 {
		delete(gs.groups, key)
	}
}

func (gs *GroupStore) Get(name string) (*Group, bool) {
	g, ok := gs.groups[groupKey(name)]
	if !ok {
		return nil, false
	}
	copy := &Group{Name: g.Name, Ports: append([]int{}, g.Ports...)}
	return copy, true
}

func (gs *GroupStore) All() []*Group {
	out := make([]*Group, 0, len(gs.groups))
	for _, g := range gs.groups {
		out = append(out, &Group{Name: g.Name, Ports: append([]int{}, g.Ports...)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (gs *GroupStore) Save() error {
	if err := os.MkdirAll(filepath.Dir(gs.path), 0755); err != nil {
		return fmt.Errorf("groupstore: mkdir: %w", err)
	}
	f, err := os.Create(gs.path)
	if err != nil {
		return fmt.Errorf("groupstore: create: %w", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(gs.groups)
}

func (gs *GroupStore) load() error {
	f, err := os.Open(gs.path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(&gs.groups)
}
