package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type RiskLevel int

const (
	RiskUnknown  RiskLevel = 0
	RiskLow      RiskLevel = 1
	RiskMedium   RiskLevel = 2
	RiskHigh     RiskLevel = 3
	RiskCritical RiskLevel = 4
)

func (r RiskLevel) String() string {
	switch r {
	case RiskLow:
		return "low"
	case RiskMedium:
		return "medium"
	case RiskHigh:
		return "high"
	case RiskCritical:
		return "critical"
	default:
		return "unknown"
	}
}

func ParseRiskLevel(s string) (RiskLevel, error) {
	switch s {
	case "low":
		return RiskLow, nil
	case "medium":
		return RiskMedium, nil
	case "high":
		return RiskHigh, nil
	case "critical":
		return RiskCritical, nil
	default:
		return RiskUnknown, fmt.Errorf("unknown risk level: %q", s)
	}
}

type RiskEntry struct {
	Port     int       `json:"port"`
	Protocol string    `json:"protocol"`
	Level    RiskLevel `json:"level"`
}

func riskKey(port int, proto string) string {
	return fmt.Sprintf("%d/%s", port, proto)
}

type RiskStore struct {
	dir     string
	entries map[string]RiskEntry
}

func NewRiskStore(dir string) (*RiskStore, error) {
	rs := &RiskStore{dir: dir, entries: make(map[string]RiskEntry)}
	path := filepath.Join(dir, "risk.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return rs, nil
		}
		return nil, fmt.Errorf("risk: read: %w", err)
	}
	if err := json.Unmarshal(data, &rs.entries); err != nil {
		return nil, fmt.Errorf("risk: unmarshal: %w", err)
	}
	return rs, nil
}

func (rs *RiskStore) Set(port int, proto string, level RiskLevel) error {
	rs.entries[riskKey(port, proto)] = RiskEntry{Port: port, Protocol: proto, Level: level}
	return rs.save()
}

func (rs *RiskStore) Get(port int, proto string) (RiskLevel, bool) {
	e, ok := rs.entries[riskKey(port, proto)]
	if !ok {
		return RiskUnknown, false
	}
	return e.Level, true
}

func (rs *RiskStore) Remove(port int, proto string) error {
	key := riskKey(port, proto)
	if _, ok := rs.entries[key]; !ok {
		return fmt.Errorf("risk: entry not found: %s", key)
	}
	delete(rs.entries, key)
	return rs.save()
}

func (rs *RiskStore) All() []RiskEntry {
	out := make([]RiskEntry, 0, len(rs.entries))
	for _, e := range rs.entries {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Port != out[j].Port {
			return out[i].Port < out[j].Port
		}
		return out[i].Protocol < out[j].Protocol
	})
	return out
}

func (rs *RiskStore) save() error {
	if err := os.MkdirAll(rs.dir, 0o755); err != nil {
		return fmt.Errorf("risk: mkdir: %w", err)
	}
	data, err := json.MarshalIndent(rs.entries, "", "  ")
	if err != nil {
		return fmt.Errorf("risk: marshal: %w", err)
	}
	return os.WriteFile(filepath.Join(rs.dir, "risk.json"), data, 0o644)
}
