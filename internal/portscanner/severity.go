package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// SeverityLevel represents how critical an unexpected port listener is.
type SeverityLevel string

const (
	SeverityLow      SeverityLevel = "low"
	SeverityMedium   SeverityLevel = "medium"
	SeverityHigh     SeverityLevel = "high"
	SeverityCritical SeverityLevel = "critical"
)

// SeverityRule maps a port+protocol combination to a severity level.
type SeverityRule struct {
	Port     int           `json:"port"`
	Protocol string        `json:"protocol"`
	Level    SeverityLevel `json:"level"`
}

// SeverityStore holds user-defined severity rules for ports.
type SeverityStore struct {
	mu    sync.RWMutex
	rules map[string]SeverityRule
	path  string
}

func severityKey(port int, protocol string) string {
	return fmt.Sprintf("%s:%d", protocol, port)
}

// NewSeverityStore creates or loads a SeverityStore from the given file path.
func NewSeverityStore(path string) (*SeverityStore, error) {
	s := &SeverityStore{
		rules: make(map[string]SeverityRule),
		path:  path,
	}
	if _, err := os.Stat(path); err == nil {
		if err := s.load(); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// Set assigns a severity level to a port/protocol pair and persists it.
func (s *SeverityStore) Set(port int, protocol string, level SeverityLevel) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rules[severityKey(port, protocol)] = SeverityRule{Port: port, Protocol: protocol, Level: level}
	return s.save()
}

// Get returns the severity level for a port/protocol pair.
// Returns SeverityLow and false if not found.
func (s *SeverityStore) Get(port int, protocol string) (SeverityLevel, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.rules[severityKey(port, protocol)]
	if !ok {
		return SeverityLow, false
	}
	return r.Level, true
}

// Remove deletes a severity rule and persists the change.
func (s *SeverityStore) Remove(port int, protocol string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.rules, severityKey(port, protocol))
	return s.save()
}

// All returns a copy of all severity rules.
func (s *SeverityStore) All() []SeverityRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]SeverityRule, 0, len(s.rules))
	for _, r := range s.rules {
		out = append(out, r)
	}
	return out
}

func (s *SeverityStore) save() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(s.path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(s.rules)
}

func (s *SeverityStore) load() error {
	f, err := os.Open(s.path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(&s.rules)
}
