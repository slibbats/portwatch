package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Snapshot represents a point-in-time capture of active listeners.
type Snapshot struct {
	Timestamp time.Time  `json:"timestamp"`
	Listeners []Listener `json:"listeners"`
}

// NewSnapshot creates a Snapshot from the current listeners.
func NewSnapshot(listeners []Listener) Snapshot {
	return Snapshot{
		Timestamp: time.Now().UTC(),
		Listeners: listeners,
	}
}

// SaveSnapshot writes a snapshot to a JSON file at the given path.
func SaveSnapshot(path string, snap Snapshot) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("snapshot: mkdir: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("snapshot: create: %w", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(snap); err != nil {
		return fmt.Errorf("snapshot: encode: %w", err)
	}
	return nil
}

// LoadSnapshot reads a snapshot from the given JSON file path.
func LoadSnapshot(path string) (Snapshot, error) {
	f, err := os.Open(path)
	if err != nil {
		return Snapshot{}, fmt.Errorf("snapshot: open: %w", err)
	}
	defer f.Close()
	var snap Snapshot
	if err := json.NewDecoder(f).Decode(&snap); err != nil {
		return Snapshot{}, fmt.Errorf("snapshot: decode: %w", err)
	}
	return snap, nil
}

// Diff returns listeners present in s but absent in other (new listeners).
func (s Snapshot) Diff(other Snapshot) []Listener {
	existing := make(map[string]struct{}, len(other.Listeners))
	for _, l := range other.Listeners {
		existing[listenerKey(l)] = struct{}{}
	}
	var added []Listener
	for _, l := range s.Listeners {
		if _, found := existing[listenerKey(l)]; !found {
			added = append(added, l)
		}
	}
	return added
}
