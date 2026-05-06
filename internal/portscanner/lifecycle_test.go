package portscanner

import (
	"os"
	"testing"
	"time"
)

func makeLifecycleListener(port int, proto string) Listener {
	return Listener{
		Port:     port,
		Protocol: proto,
		Address:  "0.0.0.0",
		Process:  "testd",
	}
}

func TestLifecycleStore_Record_And_All(t *testing.T) {
	dir := t.TempDir()
	store := NewLifecycleStore(dir)

	l := makeLifecycleListener(8080, "tcp")
	if err := store.Record(l, LifecycleOpened); err != nil {
		t.Fatalf("Record opened: %v", err)
	}
	time.Sleep(2 * time.Millisecond)
	if err := store.Record(l, LifecycleClosed); err != nil {
		t.Fatalf("Record closed: %v", err)
	}

	entries, err := store.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Event != LifecycleOpened {
		t.Errorf("expected first event=opened, got %s", entries[0].Event)
	}
	if entries[1].Event != LifecycleClosed {
		t.Errorf("expected second event=closed, got %s", entries[1].Event)
	}
}

func TestLifecycleStore_All_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	store := NewLifecycleStore(dir)
	entries, err := store.All()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestLifecycleStore_FilterByPort(t *testing.T) {
	dir := t.TempDir()
	store := NewLifecycleStore(dir)

	store.Record(makeLifecycleListener(443, "tcp"), LifecycleOpened)
	time.Sleep(2 * time.Millisecond)
	store.Record(makeLifecycleListener(8080, "tcp"), LifecycleOpened)
	time.Sleep(2 * time.Millisecond)
	store.Record(makeLifecycleListener(443, "tcp"), LifecycleClosed)

	results, err := store.FilterByPort(443, "tcp")
	if err != nil {
		t.Fatalf("FilterByPort: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 entries for port 443, got %d", len(results))
	}
	for _, e := range results {
		if e.Port != 443 || e.Protocol != "tcp" {
			t.Errorf("unexpected entry: %+v", e)
		}
	}
}

func TestLifecycleStore_Record_CreatesDir(t *testing.T) {
	base := t.TempDir()
	dir := base + "/nested/lifecycle"
	store := NewLifecycleStore(dir)

	if err := store.Record(makeLifecycleListener(22, "tcp"), LifecycleOpened); err != nil {
		t.Fatalf("Record: %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("expected dir to be created: %v", err)
	}
}

func TestLifecycleStore_All_SortedByTimestamp(t *testing.T) {
	dir := t.TempDir()
	store := NewLifecycleStore(dir)

	ports := []int{9000, 9001, 9002}
	for _, p := range ports {
		store.Record(makeLifecycleListener(p, "tcp"), LifecycleOpened)
		time.Sleep(2 * time.Millisecond)
	}

	entries, err := store.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	for i := 1; i < len(entries); i++ {
		if entries[i].Timestamp.Before(entries[i-1].Timestamp) {
			t.Errorf("entries not sorted at index %d", i)
		}
	}
}
