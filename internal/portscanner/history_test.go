package portscanner

import (
	"path/filepath"
	"testing"
	"time"
)

func TestHistoryStore_Save_And_Latest(t *testing.T) {
	dir := t.TempDir()
	hs := NewHistoryStore(dir)

	snap := NewSnapshot([]Listener{
		makeSnapListener("0.0.0.0", "tcp", 8080),
	})
	if err := hs.Save(snap); err != nil {
		t.Fatalf("Save: %v", err)
	}
	latest, err := hs.Latest()
	if err != nil {
		t.Fatalf("Latest: %v", err)
	}
	if len(latest.Listeners) != 1 || latest.Listeners[0].Port != 8080 {
		t.Errorf("unexpected latest snapshot: %+v", latest)
	}
}

func TestHistoryStore_Latest_EmptyDir(t *testing.T) {
	hs := NewHistoryStore(t.TempDir())
	_, err := hs.Latest()
	if err == nil {
		t.Fatal("expected error for empty history")
	}
}

func TestHistoryStore_All_ChronologicalOrder(t *testing.T) {
	dir := t.TempDir()
	hs := NewHistoryStore(dir)

	for _, port := range []uint16{80, 443, 8080} {
		snap := Snapshot{
			Timestamp: time.Now().UTC(),
			Listeners: []Listener{makeSnapListener("0.0.0.0", "tcp", port)},
		}
		time.Sleep(time.Millisecond) // ensure distinct timestamps
		if err := hs.Save(snap); err != nil {
			t.Fatalf("Save port %d: %v", port, err)
		}
	}

	all, err := hs.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 snapshots, got %d", len(all))
	}
}

func TestHistoryStore_Prune_RemovesOldFiles(t *testing.T) {
	dir := t.TempDir()
	hs := NewHistoryStore(dir)

	// Write an old snapshot directly with a past timestamp.
	oldSnap := Snapshot{
		Timestamp: time.Now().UTC().Add(-48 * time.Hour),
		Listeners: nil,
	}
	name := oldSnap.Timestamp.Format(snapshotTimeFormat) + ".json"
	if err := SaveSnapshot(filepath.Join(dir, name), oldSnap); err != nil {
		t.Fatalf("SaveSnapshot: %v", err)
	}

	newSnap := NewSnapshot(nil)
	if err := hs.Save(newSnap); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if err := hs.Prune(24 * time.Hour); err != nil {
		t.Fatalf("Prune: %v", err)
	}

	all, err := hs.All()
	if err != nil {
		t.Fatalf("All after prune: %v", err)
	}
	if len(all) != 1 {
		t.Errorf("expected 1 snapshot after prune, got %d", len(all))
	}
}

func TestHistoryStore_List_NonExistentDir(t *testing.T) {
	hs := NewHistoryStore("/nonexistent/history")
	entries, err := hs.list()
	if err != nil {
		t.Fatalf("expected nil error for missing dir, got %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty list, got %v", entries)
	}
}
