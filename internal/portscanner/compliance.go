package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type ComplianceStatus string

const (
	CompliancePass    ComplianceStatus = "pass"
	ComplianceFail    ComplianceStatus = "fail"
	ComplianceWarning ComplianceStatus = "warning"
	ComplianceUnknown ComplianceStatus = "unknown"
)

func ParseComplianceStatus(s string) (ComplianceStatus, error) {
	switch ComplianceStatus(s) {
	case CompliancePass, ComplianceFail, ComplianceWarning, ComplianceUnknown:
		return ComplianceStatus(s), nil
	}
	return "", fmt.Errorf("unknown compliance status: %q", s)
}

type ComplianceEntry struct {
	Port     int              `json:"port"`
	Proto    string           `json:"proto"`
	Status   ComplianceStatus `json:"status"`
	Policy   string           `json:"policy"`
	Reason   string           `json:"reason,omitempty"`
}

func complianceKey(port int, proto string) string {
	return fmt.Sprintf("%d/%s", port, proto)
}

type ComplianceStore struct {
	dir     string
	entries map[string]ComplianceEntry
}

func NewComplianceStore(dir string) (*ComplianceStore, error) {
	cs := &ComplianceStore{dir: dir, entries: make(map[string]ComplianceEntry)}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("compliance store: mkdir: %w", err)
	}
	_ = cs.load()
	return cs, nil
}

func (cs *ComplianceStore) Set(port int, proto string, status ComplianceStatus, policy, reason string) error {
	cs.entries[complianceKey(port, proto)] = ComplianceEntry{
		Port: port, Proto: proto, Status: status, Policy: policy, Reason: reason,
	}
	return cs.save()
}

func (cs *ComplianceStore) Get(port int, proto string) (ComplianceEntry, bool) {
	e, ok := cs.entries[complianceKey(port, proto)]
	return e, ok
}

func (cs *ComplianceStore) Remove(port int, proto string) error {
	key := complianceKey(port, proto)
	if _, ok := cs.entries[key]; !ok {
		return fmt.Errorf("compliance: no entry for %s", key)
	}
	delete(cs.entries, key)
	return cs.save()
}

func (cs *ComplianceStore) All() []ComplianceEntry {
	out := make([]ComplianceEntry, 0, len(cs.entries))
	for _, e := range cs.entries {
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

func (cs *ComplianceStore) save() error {
	path := filepath.Join(cs.dir, "compliance.json")
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("compliance store: save: %w", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(cs.entries)
}

func (cs *ComplianceStore) load() error {
	path := filepath.Join(cs.dir, "compliance.json")
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(&cs.entries)
}
