package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LabelStore maps port+protocol pairs to human-readable labels.
type LabelStore struct {
	path   string
	labels map[string]string
}

// labelKey returns a consistent key for a port/protocol pair.
func labelKey(port uint16, proto string) string {
	return fmt.Sprintf("%d/%s", port, proto)
}

// NewLabelStore creates or loads a LabelStore from the given directory.
func NewLabelStore(dir string) (*LabelStore, error) {
	path := filepath.Join(dir, "labels.json")
	ls := &LabelStore{path: path, labels: make(map[string]string)}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ls, nil
		}
		return nil, fmt.Errorf("label store: read: %w", err)
	}
	if err := json.Unmarshal(data, &ls.labels); err != nil {
		return nil, fmt.Errorf("label store: parse: %w", err)
	}
	return ls, nil
}

// Set assigns a label to a port/protocol pair.
func (ls *LabelStore) Set(port uint16, proto, label string) {
	ls.labels[labelKey(port, proto)] = label
}

// Get retrieves the label for a port/protocol pair, returning empty string if not found.
func (ls *LabelStore) Get(port uint16, proto string) string {
	return ls.labels[labelKey(port, proto)]
}

// Remove deletes the label for a port/protocol pair.
func (ls *LabelStore) Remove(port uint16, proto string) {
	delete(ls.labels, labelKey(port, proto))
}

// All returns a copy of all labels keyed by "port/proto".
func (ls *LabelStore) All() map[string]string {
	copy := make(map[string]string, len(ls.labels))
	for k, v := range ls.labels {
		copy[k] = v
	}
	return copy
}

// Save persists the label store to disk.
func (ls *LabelStore) Save() error {
	if err := os.MkdirAll(filepath.Dir(ls.path), 0o755); err != nil {
		return fmt.Errorf("label store: mkdir: %w", err)
	}
	data, err := json.MarshalIndent(ls.labels, "", "  ")
	if err != nil {
		return fmt.Errorf("label store: marshal: %w", err)
	}
	return os.WriteFile(ls.path, data, 0o644)
}
