package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type ignoreKey struct {
	Port     int
	Protocol string
}

// IgnoreStore holds a set of port+protocol pairs that should be silently ignored
// during scanning and alerting.
type IgnoreStore struct {
	mu      sync.RWMutex
	entries map[ignoreKey]struct{}
	path    string
}

// NewIgnoreStore creates an IgnoreStore backed by the given file path.
func NewIgnoreStore(path string) *IgnoreStore {
	return &IgnoreStore{
		entries: make(map[ignoreKey]struct{}),
		path:    path,
	}
}

// Add marks a port/protocol pair as ignored.
func (s *IgnoreStore) Add(port int, protocol string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[ignoreKey{Port: port, Protocol: protocol}] = struct{}{}
}

// Remove un-ignores a port/protocol pair.
func (s *IgnoreStore) Remove(port int, protocol string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, ignoreKey{Port: port, Protocol: protocol})
}

// Contains returns true if the given port/protocol pair is ignored.
func (s *IgnoreStore) Contains(port int, protocol string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.entries[ignoreKey{Port: port, Protocol: protocol}]
	return ok
}

// FilterIgnored returns only the listeners that are NOT in the ignore list.
func (s *IgnoreStore) FilterIgnored(listeners []Listener) []Listener {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Listener, 0, len(listeners))
	for _, l := range listeners {
		if _, ignored := s.entries[ignoreKey{Port: l.Port, Protocol: l.Protocol}]; !ignored {
			out = append(out, l)
		}
	}
	return out
}

// All returns a copy of all ignored port/protocol pairs as "protocol:port" strings.
func (s *IgnoreStore) All() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]string, 0, len(s.entries))
	for k := range s.entries {
		result = append(result, fmt.Sprintf("%s:%d", k.Protocol, k.Port))
	}
	return result
}

// Save persists the ignore list to disk.
func (s *IgnoreStore) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	type entry struct {
		Port     int    `json:"port"`
		Protocol string `json:"protocol"`
	}
	list := make([]entry, 0, len(s.entries))
	for k := range s.entries {
		list = append(list, entry{Port: k.Port, Protocol: k.Protocol})
	}
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return fmt.Errorf("ignore: marshal: %w", err)
	}
	return os.WriteFile(s.path, data, 0644)
}

// Load reads the ignore list from disk, replacing any current entries.
func (s *IgnoreStore) Load() error {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("ignore: read: %w", err)
	}
	type entry struct {
		Port     int    `json:"port"`
		Protocol string `json:"protocol"`
	}
	var list []entry
	if err := json.Unmarshal(data, &list); err != nil {
		return fmt.Errorf("ignore: unmarshal: %w", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = make(map[ignoreKey]struct{}, len(list))
	for _, e := range list {
		s.entries[ignoreKey{Port: e.Port, Protocol: e.Protocol}] = struct{}{}
	}
	return nil
}
