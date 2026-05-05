package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type FingerprintEntry struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Address  string `json:"address"`
	Process  string `json:"process"`
}

type Fingerprint struct {
	Entries []FingerprintEntry `json:"entries"`
}

func fingerprintKey(port int, proto string) string {
	return fmt.Sprintf("%d/%s", port, proto)
}

func NewFingerprint(listeners []Listener) *Fingerprint {
	entries := make([]FingerprintEntry, 0, len(listeners))
	for _, l := range listeners {
		entries = append(entries, FingerprintEntry{
			Port:     l.Port,
			Protocol: l.Protocol,
			Address:  l.Address,
			Process:  l.Process,
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		ki := fingerprintKey(entries[i].Port, entries[i].Protocol)
		kj := fingerprintKey(entries[j].Port, entries[j].Protocol)
		return ki < kj
	})
	return &Fingerprint{Entries: entries}
}

func SaveFingerprint(fp *Fingerprint, dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("fingerprint: mkdir: %w", err)
	}
	path := filepath.Join(dir, "fingerprint.json")
	data, err := json.MarshalIndent(fp, "", "  ")
	if err != nil {
		return fmt.Errorf("fingerprint: marshal: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("fingerprint: write: %w", err)
	}
	return nil
}

func LoadFingerprint(dir string) (*Fingerprint, error) {
	path := filepath.Join(dir, "fingerprint.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("fingerprint: read: %w", err)
	}
	var fp Fingerprint
	if err := json.Unmarshal(data, &fp); err != nil {
		return nil, fmt.Errorf("fingerprint: unmarshal: %w", err)
	}
	return &fp, nil
}

func DiffFingerprint(baseline, current *Fingerprint) (added, removed []FingerprintEntry) {
	baseMap := make(map[string]FingerprintEntry, len(baseline.Entries))
	for _, e := range baseline.Entries {
		baseMap[fingerprintKey(e.Port, e.Protocol)] = e
	}
	currMap := make(map[string]FingerprintEntry, len(current.Entries))
	for _, e := range current.Entries {
		currMap[fingerprintKey(e.Port, e.Protocol)] = e
	}
	for k, e := range currMap {
		if _, ok := baseMap[k]; !ok {
			added = append(added, e)
		}
	}
	for k, e := range baseMap {
		if _, ok := currMap[k]; !ok {
			removed = append(removed, e)
		}
	}
	sort.Slice(added, func(i, j int) bool { return added[i].Port < added[j].Port })
	sort.Slice(removed, func(i, j int) bool { return removed[i].Port < removed[j].Port })
	return added, removed
}
