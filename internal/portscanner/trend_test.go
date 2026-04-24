package portscanner

import (
	"testing"
	"time"
)

func makeTrendListener(proto, addr string, port uint16) Listener {
	return Listener{Proto: proto, Addr: addr, Port: port}
}

func makeTrendSnapshot(t time.Time, listeners []Listener) *Snapshot {
	return &Snapshot{CapturedAt: t, Listeners: listeners}
}

func TestAnalyzeTrends_SingleSnapshot(t *testing.T) {
	now := time.Now()
	snaps := []*Snapshot{
		makeTrendSnapshot(now, []Listener{
			makeTrendListener("tcp", "0.0.0.0", 8080),
		}),
	}
	entries := AnalyzeTrends(snaps)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].SeenIn != 1 {
		t.Errorf("expected SeenIn=1, got %d", entries[0].SeenIn)
	}
	if entries[0].TotalSnapshots != 1 {
		t.Errorf("expected TotalSnapshots=1, got %d", entries[0].TotalSnapshots)
	}
}

func TestAnalyzeTrends_PortAppearsInAll(t *testing.T) {
	base := time.Now()
	l := makeTrendListener("tcp", "0.0.0.0", 443)
	snaps := []*Snapshot{
		makeTrendSnapshot(base, []Listener{l}),
		makeTrendSnapshot(base.Add(time.Minute), []Listener{l}),
		makeTrendSnapshot(base.Add(2*time.Minute), []Listener{l}),
	}
	entries := AnalyzeTrends(snaps)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	e := entries[0]
	if e.SeenIn != 3 {
		t.Errorf("expected SeenIn=3, got %d", e.SeenIn)
	}
	if e.FrequencyRatio() != 1.0 {
		t.Errorf("expected ratio 1.0, got %f", e.FrequencyRatio())
	}
}

func TestAnalyzeTrends_PortAppearsOnce(t *testing.T) {
	base := time.Now()
	permanent := makeTrendListener("tcp", "0.0.0.0", 22)
	transient := makeTrendListener("tcp", "0.0.0.0", 9999)
	snaps := []*Snapshot{
		makeTrendSnapshot(base, []Listener{permanent, transient}),
		makeTrendSnapshot(base.Add(time.Minute), []Listener{permanent}),
		makeTrendSnapshot(base.Add(2*time.Minute), []Listener{permanent}),
	}
	entries := AnalyzeTrends(snaps)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	for _, e := range entries {
		if e.Port == 9999 && e.SeenIn != 1 {
			t.Errorf("transient port: expected SeenIn=1, got %d", e.SeenIn)
		}
		if e.Port == 22 && e.SeenIn != 3 {
			t.Errorf("permanent port: expected SeenIn=3, got %d", e.SeenIn)
		}
	}
}

func TestAnalyzeTrends_EmptySnapshots(t *testing.T) {
	entries := AnalyzeTrends([]*Snapshot{})
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for empty input, got %d", len(entries))
	}
}

func TestTrendEntry_String(t *testing.T) {
	e := TrendEntry{Proto: "tcp", Addr: "0.0.0.0", Port: 80, SeenIn: 3, TotalSnapshots: 4}
	s := e.String()
	if s == "" {
		t.Error("expected non-empty string from TrendEntry.String()")
	}
}
