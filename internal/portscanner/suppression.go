package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func suppressionKey(port int, proto string) string {
	return fmt.Sprintf("%d/%s", port, proto)
}

// SuppressionEntry represents a suppressed port alert.
type SuppressionEntry struct {
	Port      int       `json:"port"`
	Proto     string    `json:"proto"`
	Reason    string    `json:"reason"`
	Until     time.Time `json:"until"`
	CreatedAt time.Time `json:"created_at"`
}

// IsActive returns true if the suppression is still in effect.
func (e SuppressionEntry) IsActive() bool {
	return time.Now().Before(e.Until)
}

// SuppressionStore manages alert suppressions for ports.
type SuppressionStore struct {
	dir   string
	items map[string]SuppressionEntry
}

// NewSuppressionStore loads or initialises a suppression store at dir.
func NewSuppressionStore(dir string) (*SuppressionStore, error) {
	s := &SuppressionStore{dir: dir, items: make(map[string]SuppressionEntry)}
	path := filepath.Join(dir, "suppressions.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, fmt.Errorf("suppression: read: %w", err)
	}
	if err := json.Unmarshal(data, &s.items); err != nil {
		return nil, fmt.Errorf("suppression: unmarshal: %w", err)
	}
	return s, nil
}

func (s *SuppressionStore) save() error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.dir, "suppressions.json"), data, 0o644)
}

// Set adds or replaces a suppression entry.
func (s *SuppressionStore) Set(port int, proto, reason string, until time.Time) error {
	s.items[suppressionKey(port, proto)] = SuppressionEntry{
		Port:      port,
		Proto:     proto,
		Reason:    reason,
		Until:     until,
		CreatedAt: time.Now(),
	}
	return s.save()
}

// Get returns the suppression entry for the given port/proto, if present.
func (s *SuppressionStore) Get(port int, proto string) (SuppressionEntry, bool) {
	e, ok := s.items[suppressionKey(port, proto)]
	return e, ok
}

// IsSuppressed returns true when an active suppression exists for port/proto.
func (s *SuppressionStore) IsSuppressed(port int, proto string) bool {
	e, ok := s.Get(port, proto)
	return ok && e.IsActive()
}

// Remove deletes a suppression entry.
func (s *SuppressionStore) Remove(port int, proto string) error {
	key := suppressionKey(port, proto)
	if _, ok := s.items[key]; !ok {
		return fmt.Errorf("suppression: %s not found", key)
	}
	delete(s.items, key)
	return s.save()
}

// All returns all suppression entries sorted by port.
func (s *SuppressionStore) All() []SuppressionEntry {
	out := make([]SuppressionEntry, 0, len(s.items))
	for _, e := range s.items {
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

// PruneExpired removes all entries whose Until time has passed.
func (s *SuppressionStore) PruneExpired() error {
	for k, e := range s.items {
		if !e.IsActive() {
			delete(s.items, k)
		}
	}
	return s.save()
}
