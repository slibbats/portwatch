package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Baseline represents a snapshot of port listeners at a given time.
type Baseline struct {
	CapturedAt time.Time  `json:"captured_at"`
	Listeners  []Listener `json:"listeners"`
}

// Listener represents a single listening port entry.
type Listener struct {
	Addr     string `json:"addr"`
	Port     uint16 `json:"port"`
	Protocol string `json:"protocol"`
}

// CaptureBaseline scans current listeners and returns a Baseline snapshot.
func CaptureBaseline() (*Baseline, error) {
	listeners, err := ScanListeners()
	if err != nil {
		return nil, fmt.Errorf("capture baseline: %w", err)
	}
	return &Baseline{
		CapturedAt: time.Now().UTC(),
		Listeners:  listeners,
	}, nil
}

// SaveBaseline writes a Baseline snapshot to the given file path as JSON.
func SaveBaseline(path string, b *Baseline) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("save baseline: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(b); err != nil {
		return fmt.Errorf("save baseline encode: %w", err)
	}
	return nil
}

// LoadBaseline reads a Baseline snapshot from the given file path.
func LoadBaseline(path string) (*Baseline, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("load baseline: %w", err)
	}
	defer f.Close()

	var b Baseline
	if err := json.NewDecoder(f).Decode(&b); err != nil {
		return nil, fmt.Errorf("load baseline decode: %w", err)
	}
	return &b, nil
}

// Diff returns listeners present in current but absent in the baseline.
func (b *Baseline) Diff(current []Listener) []Listener {
	known := make(map[string]struct{}, len(b.Listeners))
	for _, l := range b.Listeners {
		known[listenerKey(l)] = struct{}{}
	}

	var novel []Listener
	for _, l := range current {
		if _, ok := known[listenerKey(l)]; !ok {
			novel = append(novel, l)
		}
	}
	return novel
}

func listenerKey(l Listener) string {
	return fmt.Sprintf("%s:%d/%s", l.Addr, l.Port, l.Protocol)
}
