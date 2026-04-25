package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// Tag associates a human-readable label with a port/protocol pair.
type Tag struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Label    string `json:"label"`
	Note     string `json:"note,omitempty"`
}

// TagStore holds a collection of port tags.
type TagStore struct {
	Tags []Tag `json:"tags"`
}

// NewTagStore returns an empty TagStore.
func NewTagStore() *TagStore {
	return &TagStore{Tags: []Tag{}}
}

// Add inserts or updates a tag for the given port/protocol pair.
func (ts *TagStore) Add(t Tag) {
	for i, existing := range ts.Tags {
		if existing.Port == t.Port && existing.Protocol == t.Protocol {
			ts.Tags[i] = t
			return
		}
	}
	ts.Tags = append(ts.Tags, t)
}

// Get returns the tag for a port/protocol pair, if any.
func (ts *TagStore) Get(port int, protocol string) (Tag, bool) {
	for _, t := range ts.Tags {
		if t.Port == port && t.Protocol == protocol {
			return t, true
		}
	}
	return Tag{}, false
}

// Remove deletes a tag by port/protocol pair.
func (ts *TagStore) Remove(port int, protocol string) bool {
	for i, t := range ts.Tags {
		if t.Port == port && t.Protocol == protocol {
			ts.Tags = append(ts.Tags[:i], ts.Tags[i+1:]...)
			return true
		}
	}
	return false
}

// Sorted returns a copy of tags sorted by port then protocol.
func (ts *TagStore) Sorted() []Tag {
	copy := append([]Tag{}, ts.Tags...)
	sort.Slice(copy, func(i, j int) bool {
		if copy[i].Port != copy[j].Port {
			return copy[i].Port < copy[j].Port
		}
		return copy[i].Protocol < copy[j].Protocol
	})
	return copy
}

// SaveTagStore writes the tag store to a JSON file.
func SaveTagStore(path string, ts *TagStore) error {
	data, err := json.MarshalIndent(ts, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal tag store: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// LoadTagStore reads a tag store from a JSON file.
func LoadTagStore(path string) (*TagStore, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return NewTagStore(), nil
		}
		return nil, fmt.Errorf("read tag store: %w", err)
	}
	var ts TagStore
	if err := json.Unmarshal(data, &ts); err != nil {
		return nil, fmt.Errorf("unmarshal tag store: %w", err)
	}
	return &ts, nil
}
