package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type RateLimitEntry struct {
	Port     int       `json:"port"`
	Proto    string    `json:"proto"`
	MaxPerHour int     `json:"max_per_hour"`
	CreatedAt  time.Time `json:"created_at"`
}

type RateLimitStore struct {
	mu      sync.RWMutex
	entries map[string]RateLimitEntry
	path    string
}

func rateLimitKey(port int, proto string) string {
	return fmt.Sprintf("%d/%s", port, proto)
}

func NewRateLimitStore(dir string) (*RateLimitStore, error) {
	s := &RateLimitStore{
		entries: make(map[string]RateLimitEntry),
		path:    filepath.Join(dir, "ratelimits.json"),
	}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return s, nil
}

func (s *RateLimitStore) Set(port int, proto string, maxPerHour int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[rateLimitKey(port, proto)] = RateLimitEntry{
		Port:       port,
		Proto:      proto,
		MaxPerHour: maxPerHour,
		CreatedAt:  time.Now(),
	}
	return s.save()
}

func (s *RateLimitStore) Get(port int, proto string) (RateLimitEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[rateLimitKey(port, proto)]
	return e, ok
}

func (s *RateLimitStore) Remove(port int, proto string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := rateLimitKey(port, proto)
	if _, ok := s.entries[key]; !ok {
		return fmt.Errorf("no rate limit entry for %s", key)
	}
	delete(s.entries, key)
	return s.save()
}

func (s *RateLimitStore) All() []RateLimitEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]RateLimitEntry, 0, len(s.entries))
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

func (s *RateLimitStore) save() error {
	data, err := json.MarshalIndent(s.entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

func (s *RateLimitStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.entries)
}
