package portscanner

import (
	"testing"
	"time"
)

func makeAnomalyListener(ip, proto string, port uint16) Listener {
	return Listener{IP: ip, Port: port, Protocol: proto}
}

func makeAnomalySnapshot(listeners []Listener) *Snapshot {
	return &Snapshot{
		Timestamp: time.Now(),
		Listeners: listeners,
	}
}

func TestDetectAnomalies_TooFewSnapshots(t *testing.T) {
	result := DetectAnomalies([]*Snapshot{makeAnomalySnapshot(nil)})
	if len(result) != 0 {
		t.Errorf("expected no anomalies with single snapshot, got %d", len(result))
	}
}

func TestDetectAnomalies_NewPort(t *testing.T) {
	l := makeAnomalyListener("0.0.0.0", "tcp", 9090)
	snaps := []*Snapshot{
		makeAnomalySnapshot([]Listener{}),
		makeAnomalySnapshot([]Listener{l}),
	}

	anomalies := DetectAnomalies(snaps)
	if len(anomalies) != 1 {
		t.Fatalf("expected 1 anomaly, got %d", len(anomalies))
	}
	if anomalies[0].Kind != AnomalyNewPort {
		t.Errorf("expected AnomalyNewPort, got %s", anomalies[0].Kind)
	}
	if anomalies[0].Listener.Port != 9090 {
		t.Errorf("expected port 9090, got %d", anomalies[0].Listener.Port)
	}
}

func TestDetectAnomalies_RemovedPort(t *testing.T) {
	l := makeAnomalyListener("0.0.0.0", "tcp", 8080)
	snaps := []*Snapshot{
		makeAnomalySnapshot([]Listener{l}),
		makeAnomalySnapshot([]Listener{}),
	}

	anomalies := DetectAnomalies(snaps)
	if len(anomalies) != 1 {
		t.Fatalf("expected 1 anomaly, got %d", len(anomalies))
	}
	if anomalies[0].Kind != AnomalyRemovedPort {
		t.Errorf("expected AnomalyRemovedPort, got %s", anomalies[0].Kind)
	}
}

func TestDetectAnomalies_NoChanges(t *testing.T) {
	l := makeAnomalyListener("127.0.0.1", "tcp", 443)
	snaps := []*Snapshot{
		makeAnomalySnapshot([]Listener{l}),
		makeAnomalySnapshot([]Listener{l}),
	}

	anomalies := DetectAnomalies(snaps)
	for _, a := range anomalies {
		if a.Kind == AnomalyNewPort || a.Kind == AnomalyRemovedPort {
			t.Errorf("unexpected anomaly %s for stable listener", a.Kind)
		}
	}
}

func TestDetectAnomalies_FlappingPort(t *testing.T) {
	l := makeAnomalyListener("0.0.0.0", "udp", 5353)
	empty := []Listener{}
	snaps := []*Snapshot{
		makeAnomalySnapshot([]Listener{l}),
		makeAnomalySnapshot(empty),
		makeAnomalySnapshot([]Listener{l}),
		makeAnomalySnapshot(empty),
	}

	anomalies := DetectAnomalies(snaps)
	found := false
	for _, a := range anomalies {
		if a.Kind == AnomalyFlapping {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected a flapping anomaly to be detected")
	}
}

func TestAnomaly_String(t *testing.T) {
	a := Anomaly{
		Kind:     AnomalyNewPort,
		Listener: makeAnomalyListener("0.0.0.0", "tcp", 8080),
		Message:  "port appeared since last scan",
	}
	s := a.String()
	if s == "" {
		t.Error("expected non-empty string from Anomaly.String()")
	}
}
