package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type ScheduleEntry struct {
	Port     int       `json:"port"`
	Proto    string    `json:"proto"`
	Cron     string    `json:"cron"`
	Label    string    `json:"label,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func scheduleKey(port int, proto string) string {
	return fmt.Sprintf("%d/%s", port, proto)
}

type ScheduleStore struct {
	path    string
	entries map[string]ScheduleEntry
}

func NewScheduleStore(dir string) (*ScheduleStore, error) {
	s := &ScheduleStore{
		path:    filepath.Join(dir, "schedules.json"),
		entries: make(map[string]ScheduleEntry),
	}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return s, nil
}

func (s *ScheduleStore) Set(port int, proto, cron, label string) error {
	s.entries[scheduleKey(port, proto)] = ScheduleEntry{
		Port:      port,
		Proto:     proto,
		Cron:      cron,
		Label:     label,
		CreatedAt: time.Now().UTC(),
	}
	return s.save()
}

func (s *ScheduleStore) Get(port int, proto string) (ScheduleEntry, bool) {
	e, ok := s.entries[scheduleKey(port, proto)]
	return e, ok
}

func (s *ScheduleStore) Remove(port int, proto string) error {
	key := scheduleKey(port, proto)
	if _, ok := s.entries[key]; !ok {
		return fmt.Errorf("no schedule entry for %s", key)
	}
	delete(s.entries, key)
	return s.save()
}

func (s *ScheduleStore) All() []ScheduleEntry {
	out := make([]ScheduleEntry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Port != out[j].Port {
			return out[i].Port < out[j].Port
		}
		return out[i].Proto < out[j].Proto
	})
	return out
}

func (s *ScheduleStore) save() error {
	data, err := json.MarshalIndent(s.entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

func (s *ScheduleStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.entries)
}
