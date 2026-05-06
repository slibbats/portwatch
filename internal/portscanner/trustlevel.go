package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type TrustLevel int

const (
	TrustUnknown   TrustLevel = iota
	TrustUntrusted            // 1
	TrustLow                  // 2
	TrustMedium               // 3
	TrustHigh                 // 4
	TrustVerified             // 5
)

func (t TrustLevel) String() string {
	switch t {
	case TrustUntrusted:
		return "untrusted"
	case TrustLow:
		return "low"
	case TrustMedium:
		return "medium"
	case TrustHigh:
		return "high"
	case TrustVerified:
		return "verified"
	default:
		return "unknown"
	}
}

func ParseTrustLevel(s string) (TrustLevel, error) {
	switch s {
	case "untrusted":
		return TrustUntrusted, nil
	case "low":
		return TrustLow, nil
	case "medium":
		return TrustMedium, nil
	case "high":
		return TrustHigh, nil
	case "verified":
		return TrustVerified, nil
	default:
		return TrustUnknown, fmt.Errorf("unknown trust level: %q", s)
	}
}

type trustEntry struct {
	Port     int        `json:"port"`
	Proto    string     `json:"proto"`
	Trust    TrustLevel `json:"trust"`
}

func trustKey(port int, proto string) string {
	return fmt.Sprintf("%d/%s", port, proto)
}

type TrustLevelStore struct {
	dir     string
	entries map[string]trustEntry
}

func NewTrustLevelStore(dir string) (*TrustLevelStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("trustlevel: mkdir: %w", err)
	}
	s := &TrustLevelStore{dir: dir, entries: make(map[string]trustEntry)}
	_ = s.load()
	return s, nil
}

func (s *TrustLevelStore) Set(port int, proto string, level TrustLevel) error {
	s.entries[trustKey(port, proto)] = trustEntry{Port: port, Proto: proto, Trust: level}
	return s.save()
}

func (s *TrustLevelStore) Get(port int, proto string) (TrustLevel, bool) {
	e, ok := s.entries[trustKey(port, proto)]
	if !ok {
		return TrustUnknown, false
	}
	return e.Trust, true
}

func (s *TrustLevelStore) Remove(port int, proto string) error {
	key := trustKey(port, proto)
	if _, ok := s.entries[key]; !ok {
		return fmt.Errorf("trustlevel: entry not found: %s", key)
	}
	delete(s.entries, key)
	return s.save()
}

func (s *TrustLevelStore) All() []trustEntry {
	out := make([]trustEntry, 0, len(s.entries))
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

func (s *TrustLevelStore) file() string {
	return filepath.Join(s.dir, "trustlevel.json")
}

func (s *TrustLevelStore) save() error {
	data, err := json.MarshalIndent(s.All(), "", "  ")
	if err != nil {
		return fmt.Errorf("trustlevel: marshal: %w", err)
	}
	return os.WriteFile(s.file(), data, 0o644)
}

func (s *TrustLevelStore) load() error {
	data, err := os.ReadFile(s.file())
	if err != nil {
		return nil
	}
	var entries []trustEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("trustlevel: unmarshal: %w", err)
	}
	for _, e := range entries {
		s.entries[trustKey(e.Port, e.Proto)] = e
	}
	return nil
}
