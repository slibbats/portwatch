package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type LifecycleEvent string

const (
	LifecycleOpened LifecycleEvent = "opened"
	LifecycleClosed LifecycleEvent = "closed"
)

type LifecycleEntry struct {
	Port      int            `json:"port"`
	Protocol  string         `json:"protocol"`
	Address   string         `json:"address"`
	Process   string         `json:"process"`
	Event     LifecycleEvent `json:"event"`
	Timestamp time.Time      `json:"timestamp"`
}

func lifecycleKey(port int, proto string) string {
	return fmt.Sprintf("%d/%s", port, proto)
}

type LifecycleStore struct {
	dir string
}

func NewLifecycleStore(dir string) *LifecycleStore {
	return &LifecycleStore{dir: dir}
}

func (s *LifecycleStore) Record(l Listener, event LifecycleEvent) error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return fmt.Errorf("lifecycle: mkdir: %w", err)
	}
	entry := LifecycleEntry{
		Port:      l.Port,
		Protocol:  l.Protocol,
		Address:   l.Address,
		Process:   l.Process,
		Event:     event,
		Timestamp: time.Now().UTC(),
	}
	filename := fmt.Sprintf("%s_%s_%s.json",
		entry.Timestamp.Format("20060102T150405.000000000Z"),
		lifecycleKey(l.Port, l.Protocol),
		string(event),
	)
	path := filepath.Join(s.dir, filename)
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("lifecycle: marshal: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

func (s *LifecycleStore) All() ([]LifecycleEntry, error) {
	entries, err := filepath.Glob(filepath.Join(s.dir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("lifecycle: glob: %w", err)
	}
	var result []LifecycleEntry
	for _, path := range entries {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var e LifecycleEntry
		if err := json.Unmarshal(data, &e); err != nil {
			continue
		}
		result = append(result, e)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.Before(result[j].Timestamp)
	})
	return result, nil
}

func (s *LifecycleStore) FilterByPort(port int, proto string) ([]LifecycleEntry, error) {
	all, err := s.All()
	if err != nil {
		return nil, err
	}
	var out []LifecycleEntry
	for _, e := range all {
		if e.Port == port && e.Protocol == proto {
			out = append(out, e)
		}
	}
	return out, nil
}
