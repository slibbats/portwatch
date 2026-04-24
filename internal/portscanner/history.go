package portscanner

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const snapshotTimeFormat = "20060102T150405Z"

// HistoryStore manages a directory of timestamped snapshots.
type HistoryStore struct {
	Dir string
}

// NewHistoryStore creates a HistoryStore rooted at dir.
func NewHistoryStore(dir string) *HistoryStore {
	return &HistoryStore{Dir: dir}
}

// Save writes snap to the history directory with a timestamp-based filename.
func (h *HistoryStore) Save(snap Snapshot) error {
	name := snap.Timestamp.UTC().Format(snapshotTimeFormat) + ".json"
	return SaveSnapshot(filepath.Join(h.Dir, name), snap)
}

// Latest returns the most recent snapshot from the history directory,
// or an error if no snapshots exist.
func (h *HistoryStore) Latest() (Snapshot, error) {
	entries, err := h.list()
	if err != nil {
		return Snapshot{}, err
	}
	if len(entries) == 0 {
		return Snapshot{}, fmt.Errorf("history: no snapshots in %s", h.Dir)
	}
	return LoadSnapshot(filepath.Join(h.Dir, entries[len(entries)-1]))
}

// All returns all snapshots in chronological order.
func (h *HistoryStore) All() ([]Snapshot, error) {
	entries, err := h.list()
	if err != nil {
		return nil, err
	}
	snaps := make([]Snapshot, 0, len(entries))
	for _, e := range entries {
		s, err := LoadSnapshot(filepath.Join(h.Dir, e))
		if err != nil {
			return nil, err
		}
		snaps = append(snaps, s)
	}
	return snaps, nil
}

// Prune removes snapshots older than the given duration.
func (h *HistoryStore) Prune(maxAge time.Duration) error {
	entries, err := h.list()
	if err != nil {
		return err
	}
	cutoff := time.Now().UTC().Add(-maxAge)
	for _, e := range entries {
		ts, err := time.Parse(snapshotTimeFormat+".json", e)
		if err != nil {
			continue
		}
		if ts.Before(cutoff) {
			_ = os.Remove(filepath.Join(h.Dir, e))
		}
	}
	return nil
}

func (h *HistoryStore) list() ([]string, error) {
	entries, err := os.ReadDir(h.Dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("history: readdir: %w", err)
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}
