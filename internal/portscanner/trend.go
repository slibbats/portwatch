package portscanner

import (
	"fmt"
	"time"
)

// TrendEntry represents a port's appearance count over a time window.
type TrendEntry struct {
	Key      string
	Proto    string
	Addr     string
	Port     uint16
	SeenIn   int
	TotalSnapshots int
	FirstSeen time.Time
	LastSeen  time.Time
}

// FrequencyRatio returns the fraction of snapshots in which this port appeared.
func (t TrendEntry) FrequencyRatio() float64 {
	if t.TotalSnapshots == 0 {
		return 0
	}
	return float64(t.SeenIn) / float64(t.TotalSnapshots)
}

// String returns a human-readable summary of the trend entry.
func (t TrendEntry) String() string {
	return fmt.Sprintf("%s/%s:%d seen %d/%d snapshots (%.0f%%)",
		t.Proto, t.Addr, t.Port,
		t.SeenIn, t.TotalSnapshots,
		t.FrequencyRatio()*100,
	)
}

// AnalyzeTrends inspects a slice of snapshots and returns frequency data
// for every unique listener observed across all snapshots.
func AnalyzeTrends(snapshots []*Snapshot) []TrendEntry {
	type accumulator struct {
		entry TrendEntry
	}

	total := len(snapshots)
	acc := make(map[string]*accumulator)

	for _, snap := range snapshots {
		seen := make(map[string]bool)
		for _, l := range snap.Listeners {
			k := listenerKey(l)
			if !seen[k] {
				seen[k] = true
				if _, ok := acc[k]; !ok {
					acc[k] = &accumulator{
						entry: TrendEntry{
							Key:   k,
							Proto: l.Proto,
							Addr:  l.Addr,
							Port:  l.Port,
							FirstSeen: snap.CapturedAt,
						},
					},
				}
			}
			a := acc[k]
			a.entry.SeenIn++
			a.entry.TotalSnapshots = total
			if snap.CapturedAt.After(a.entry.LastSeen) {
				a.entry.LastSeen = snap.CapturedAt
			}
		}
	}

	results := make([]TrendEntry, 0, len(acc))
	for _, a := range acc {
		results = append(results, a.entry)
	}
	return results
}
