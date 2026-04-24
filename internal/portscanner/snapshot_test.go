package portscanner

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func makeSnapListener(addr, proto string, port uint16) Listener {
	return Listener{Addr: addr, Port: port, Protocol: proto}
}

func TestNewSnapshot_SetsTimestamp(t *testing.T) {
	before := time.Now().UTC()
	snap := NewSnapshot(nil)
	after := time.Now().UTC()
	if snap.Timestamp.Before(before) || snap.Timestamp.After(after) {
		t.Errorf("unexpected timestamp: %v", snap.Timestamp)
	}
}

func TestNewSnapshot_StoresListeners(t *testing.T) {
	listeners := []Listener{
		makeSnapListener("0.0.0.0", "tcp", 80),
		makeSnapListener("127.0.0.1", "tcp", 22),
	}
	snap := NewSnapshot(listeners)
	if len(snap.Listeners) != 2 {
		t.Fatalf("expected 2 listeners, got %d", len(snap.Listeners))
	}
}

func TestSaveAndLoadSnapshot_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "snap.json")

	original := NewSnapshot([]Listener{
		makeSnapListener("0.0.0.0", "tcp", 443),
	})
	if err := SaveSnapshot(path, original); err != nil {
		t.Fatalf("SaveSnapshot: %v", err)
	}
	loaded, err := LoadSnapshot(path)
	if err != nil {
		t.Fatalf("LoadSnapshot: %v", err)
	}
	if len(loaded.Listeners) != 1 {
		t.Fatalf("expected 1 listener, got %d", len(loaded.Listeners))
	}
	if loaded.Listeners[0].Port != 443 {
		t.Errorf("expected port 443, got %d", loaded.Listeners[0].Port)
	}
}

func TestLoadSnapshot_FileNotFound(t *testing.T) {
	_, err := LoadSnapshot("/nonexistent/path/snap.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestSnapshot_Diff_DetectsNewListeners(t *testing.T) {
	old := NewSnapshot([]Listener{
		makeSnapListener("0.0.0.0", "tcp", 80),
	})
	current := NewSnapshot([]Listener{
		makeSnapListener("0.0.0.0", "tcp", 80),
		makeSnapListener("0.0.0.0", "tcp", 9090),
	})
	added := current.Diff(old)
	if len(added) != 1 {
		t.Fatalf("expected 1 new listener, got %d", len(added))
	}
	if added[0].Port != 9090 {
		t.Errorf("expected port 9090, got %d", added[0].Port)
	}
}

func TestSnapshot_Diff_NoNewListeners(t *testing.T) {
	listeners := []Listener{makeSnapListener("0.0.0.0", "tcp", 80)}
	old := NewSnapshot(listeners)
	current := NewSnapshot(listeners)
	if diff := current.Diff(old); len(diff) != 0 {
		t.Errorf("expected no diff, got %v", diff)
	}
}

func TestSaveSnapshot_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "nested", "snap.json")
	snap := NewSnapshot(nil)
	if err := SaveSnapshot(path, snap); err != nil {
		t.Fatalf("SaveSnapshot: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file to exist: %v", err)
	}
}
