package portscanner

import "fmt"

// DiffResult holds the outcome of comparing two snapshots.
type DiffResult struct {
	New     []Listener
	Removed []Listener
}

// IsEmpty returns true when no changes were detected.
func (d DiffResult) IsEmpty() bool {
	return len(d.New) == 0 && len(d.Removed) == 0
}

// Summary returns a human-readable one-line description of the diff.
func (d DiffResult) Summary() string {
	return fmt.Sprintf("+%d new, -%d removed", len(d.New), len(d.Removed))
}

// DiffSnapshots compares two snapshots and returns listeners that were added
// or removed between the previous and current snapshot.
func DiffSnapshots(prev, curr *Snapshot) DiffResult {
	prevIndex := indexListeners(prev.Listeners)
	currIndex := indexListeners(curr.Listeners)

	var result DiffResult

	for key, l := range currIndex {
		if _, found := prevIndex[key]; !found {
			result.New = append(result.New, l)
		}
	}

	for key, l := range prevIndex {
		if _, found := currIndex[key]; !found {
			result.Removed = append(result.Removed, l)
		}
	}

	return result
}

func indexListeners(listeners []Listener) map[string]Listener {
	idx := make(map[string]Listener, len(listeners))
	for _, l := range listeners {
		idx[listenerKey(l)] = l
	}
	return idx
}
