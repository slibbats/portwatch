package portscanner

import (
	"testing"
	"time"
)

func makeDiffListener(proto, addr string, port uint16) Listener {
	return Listener{Protocol: proto, Address: addr, Port: port}
}

func makeSnapshotWith(listeners []Listener) *Snapshot {
	return &Snapshot{Timestamp: time.Now(), Listeners: listeners}
}

func TestDiffSnapshots_NoChanges(t *testing.T) {
	listeners := []Listener{
		makeDiffListener("tcp", "0.0.0.0", 8080),
		makeDiffListener("tcp", "127.0.0.1", 5432),
	}
	prev := makeSnapshotWith(listeners)
	curr := makeSnapshotWith(listeners)

	result := DiffSnapshots(prev, curr)

	if !result.IsEmpty() {
		t.Errorf("expected empty diff, got %s", result.Summary())
	}
}

func TestDiffSnapshots_DetectsNewListener(t *testing.T) {
	prev := makeSnapshotWith([]Listener{
		makeDiffListener("tcp", "0.0.0.0", 8080),
	})
	curr := makeSnapshotWith([]Listener{
		makeDiffListener("tcp", "0.0.0.0", 8080),
		makeDiffListener("tcp", "0.0.0.0", 9090),
	})

	result := DiffSnapshots(prev, curr)

	if len(result.New) != 1 {
		t.Fatalf("expected 1 new listener, got %d", len(result.New))
	}
	if result.New[0].Port != 9090 {
		t.Errorf("expected new port 9090, got %d", result.New[0].Port)
	}
	if len(result.Removed) != 0 {
		t.Errorf("expected no removed listeners, got %d", len(result.Removed))
	}
}

func TestDiffSnapshots_DetectsRemovedListener(t *testing.T) {
	prev := makeSnapshotWith([]Listener{
		makeDiffListener("tcp", "0.0.0.0", 8080),
		makeDiffListener("udp", "0.0.0.0", 53),
	})
	curr := makeSnapshotWith([]Listener{
		makeDiffListener("tcp", "0.0.0.0", 8080),
	})

	result := DiffSnapshots(prev, curr)

	if len(result.Removed) != 1 {
		t.Fatalf("expected 1 removed listener, got %d", len(result.Removed))
	}
	if result.Removed[0].Port != 53 {
		t.Errorf("expected removed port 53, got %d", result.Removed[0].Port)
	}
}

func TestDiffSnapshots_Summary(t *testing.T) {
	prev := makeSnapshotWith([]Listener{makeDiffListener("tcp", "0.0.0.0", 80)})
	curr := makeSnapshotWith([]Listener{makeDiffListener("tcp", "0.0.0.0", 443)})

	result := DiffSnapshots(prev, curr)
	summary := result.Summary()

	if summary != "+1 new, -1 removed" {
		t.Errorf("unexpected summary: %q", summary)
	}
}

func TestDiffSnapshots_BothEmpty(t *testing.T) {
	prev := makeSnapshotWith(nil)
	curr := makeSnapshotWith(nil)

	result := DiffSnapshots(prev, curr)

	if !result.IsEmpty() {
		t.Error("expected empty diff for two empty snapshots")
	}
}
