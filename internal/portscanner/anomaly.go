package portscanner

import (
	"fmt"
	"strings"
)

// AnomalyKind describes the type of anomaly detected.
type AnomalyKind string

const (
	AnomalyNewPort     AnomalyKind = "new_port"
	AnomalyRemovedPort AnomalyKind = "removed_port"
	AnomalyFlapping    AnomalyKind = "flapping"
)

// Anomaly represents a detected anomaly in port listener behaviour.
type Anomaly struct {
	Kind     AnomalyKind
	Listener Listener
	Message  string
}

func (a Anomaly) String() string {
	return fmt.Sprintf("[%s] %s:%d (%s) — %s",
		strings.ToUpper(string(a.Kind)),
		a.Listener.IP,
		a.Listener.Port,
		a.Listener.Protocol,
		a.Message,
	)
}

// DetectAnomalies compares a sequence of snapshots and returns detected anomalies.
// It flags ports that are new, removed, or flapping (appearing and disappearing).
func DetectAnomalies(snapshots []*Snapshot) []Anomaly {
	if len(snapshots) < 2 {
		return nil
	}

	var anomalies []Anomaly

	prev := indexListeners(snapshots[len(snapshots)-2].Listeners)
	curr := indexListeners(snapshots[len(snapshots)-1].Listeners)

	for key, l := range curr {
		if _, existed := prev[key]; !existed {
			anomalies = append(anomalies, Anomaly{
				Kind:     AnomalyNewPort,
				Listener: l,
				Message:  "port appeared since last scan",
			})
		}
	}

	for key, l := range prev {
		if _, stillPresent := curr[key]; !stillPresent {
			anomalies = append(anomalies, Anomaly{
				Kind:     AnomalyRemovedPort,
				Listener: l,
				Message:  "port disappeared since last scan",
			})
		}
	}

	flapping := detectFlapping(snapshots)
	anomalies = append(anomalies, flapping...)

	return anomalies
}

// detectFlapping finds ports that appear and disappear across the snapshot history.
func detectFlapping(snapshots []*Snapshot) []Anomaly {
	presence := make(map[string][]bool)

	for _, snap := range snapshots {
		seen := indexListeners(snap.Listeners)
		for key := range seen {
			presence[key] = append(presence[key], true)
		}
		for key := range presence {
			if _, ok := seen[key]; !ok {
				presence[key] = append(presence[key], false)
			}
		}
	}

	var anomalies []Anomaly
	for key, states := range presence {
		if isFlapping(states) {
			parts := strings.SplitN(key, "/", 2)
			if len(parts) != 2 {
				continue
			}
			anomalies = append(anomalies, Anomaly{
				Kind:    AnomalyFlapping,
				Message: fmt.Sprintf("port %s is flapping across %d snapshots", key, len(snapshots)),
			})
		}
	}
	return anomalies
}

func isFlapping(states []bool) bool {
	if len(states) < 3 {
		return false
	}
	transitions := 0
	for i := 1; i < len(states); i++ {
		if states[i] != states[i-1] {
			transitions++
		}
	}
	return transitions >= 2
}
