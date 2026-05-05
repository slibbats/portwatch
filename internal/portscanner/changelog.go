package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type ChangelogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Port      int       `json:"port"`
	Proto     string    `json:"proto"`
	Address   string    `json:"address"`
	Process   string    `json:"process"`
	Event     string    `json:"event"` // "added" or "removed"
}

type ChangelogStore struct {
	path string
}

func changelogKey(port int, proto string) string {
	return fmt.Sprintf("%d/%s", port, proto)
}

func NewChangelogStore(dir string) *ChangelogStore {
	return &ChangelogStore{path: filepath.Join(dir, "changelog.json")}
}

func (c *ChangelogStore) load() ([]ChangelogEntry, error) {
	data, err := os.ReadFile(c.path)
	if os.IsNotExist(err) {
		return []ChangelogEntry{}, nil
	}
	if err != nil {
		return nil, err
	}
	var entries []ChangelogEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func (c *ChangelogStore) save(entries []ChangelogEntry) error {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0o644)
}

func (c *ChangelogStore) Append(entry ChangelogEntry) error {
	entries, err := c.load()
	if err != nil {
		return err
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}
	entries = append(entries, entry)
	return c.save(entries)
}

func (c *ChangelogStore) All() ([]ChangelogEntry, error) {
	entries, err := c.load()
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.Before(entries[j].Timestamp)
	})
	return entries, nil
}

func (c *ChangelogStore) FilterByPort(port int, proto string) ([]ChangelogEntry, error) {
	all, err := c.All()
	if err != nil {
		return nil, err
	}
	key := changelogKey(port, proto)
	var result []ChangelogEntry
	for _, e := range all {
		if changelogKey(e.Port, e.Proto) == key {
			result = append(result, e)
		}
	}
	return result, nil
}

func (c *ChangelogStore) RecordDiff(diff SnapshotDiff) error {
	for _, l := range diff.Added {
		if err := c.Append(ChangelogEntry{
			Port:    l.Port,
			Proto:   l.Proto,
			Address: l.Address,
			Process: l.Process,
			Event:   "added",
		}); err != nil {
			return err
		}
	}
	for _, l := range diff.Removed {
		if err := c.Append(ChangelogEntry{
			Port:    l.Port,
			Proto:   l.Proto,
			Address: l.Address,
			Process: l.Process,
			Event:   "removed",
		}); err != nil {
			return err
		}
	}
	return nil
}
