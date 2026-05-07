package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// CorrelationEntry links a port/protocol pair to one or more related ports,
// allowing operators to document service dependencies and co-occurrence patterns.
type CorrelationEntry struct {
	Port         int       `json:"port"`
	Proto        string    `json:"proto"`
	RelatedPorts []int     `json:"related_ports"`
	Reason       string    `json:"reason,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func correlationKey(port int, proto string) string {
	return fmt.Sprintf("%d_%s", port, proto)
}

// CorrelationStore persists port correlation data to a directory.
type CorrelationStore struct {
	dir string
}

// NewCorrelationStore returns a CorrelationStore backed by dir.
func NewCorrelationStore(dir string) *CorrelationStore {
	return &CorrelationStore{dir: dir}
}

func (s *CorrelationStore) path(port int, proto string) string {
	return filepath.Join(s.dir, correlationKey(port, proto)+".json")
}

// Set stores or replaces the correlation entry for the given port/proto pair.
func (s *CorrelationStore) Set(entry CorrelationEntry) error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return fmt.Errorf("correlation: mkdir: %w", err)
	}
	now := time.Now().UTC()
	if entry.CreatedAt.IsZero() {
		existing, err := s.Get(entry.Port, entry.Proto)
		if err == nil {
			entry.CreatedAt = existing.CreatedAt
		} else {
			entry.CreatedAt = now
		}
	}
	entry.UpdatedAt = now
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("correlation: marshal: %w", err)
	}
	if err := os.WriteFile(s.path(entry.Port, entry.Proto), data, 0o644); err != nil {
		return fmt.Errorf("correlation: write: %w", err)
	}
	return nil
}

// Get retrieves the correlation entry for the given port/proto pair.
func (s *CorrelationStore) Get(port int, proto string) (CorrelationEntry, error) {
	data, err := os.ReadFile(s.path(port, proto))
	if err != nil {
		if os.IsNotExist(err) {
			return CorrelationEntry{}, fmt.Errorf("correlation: not found: %d/%s", port, proto)
		}
		return CorrelationEntry{}, fmt.Errorf("correlation: read: %w", err)
	}
	var entry CorrelationEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return CorrelationEntry{}, fmt.Errorf("correlation: unmarshal: %w", err)
	}
	return entry, nil
}

// Remove deletes the correlation entry for the given port/proto pair.
func (s *CorrelationStore) Remove(port int, proto string) error {
	err := os.Remove(s.path(port, proto))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("correlation: remove: %w", err)
	}
	return nil
}

// All returns all stored correlation entries, sorted by port then proto.
func (s *CorrelationStore) All() ([]CorrelationEntry, error) {
	entries, err := filepath.Glob(filepath.Join(s.dir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("correlation: glob: %w", err)
	}
	var result []CorrelationEntry
	for _, f := range entries {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		var entry CorrelationEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			continue
		}
		result = append(result, entry)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Port != result[j].Port {
			return result[i].Port < result[j].Port
		}
		return result[i].Proto < result[j].Proto
	})
	return result, nil
}

// RelatedFor returns the set of ports correlated with the given port/proto pair.
// Returns nil if no entry exists.
func (s *CorrelationStore) RelatedFor(port int, proto string) []int {
	entry, err := s.Get(port, proto)
	if err != nil {
		return nil
	}
	return entry.RelatedPorts
}
