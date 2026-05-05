package portscanner

import (
	"os"
	"testing"
	"time"
)

func makeChangelogListener(port int, proto string) Listener {
	return Listener{Port: port, Proto: proto, Address: "0.0.0.0", Process: "testd"}
}

func TestChangelogStore_Append_And_All(t *testing.T) {
	dir := t.TempDir()
	store := NewChangelogStore(dir)

	entry := ChangelogEntry{
		Port:  8080,
		Proto: "tcp",
		Event: "added",
	}
	if err := store.Append(entry); err != nil {
		t.Fatalf("Append: %v", err)
	}

	all, err := store.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(all))
	}
	if all[0].Port != 8080 || all[0].Event != "added" {
		t.Errorf("unexpected entry: %+v", all[0])
	}
}

func TestChangelogStore_All_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	store := NewChangelogStore(dir)

	all, err := store.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("expected 0 entries, got %d", len(all))
	}
}

func TestChangelogStore_All_SortedByTimestamp(t *testing.T) {
	dir := t.TempDir()
	store := NewChangelogStore(dir)

	now := time.Now().UTC()
	_ = store.Append(ChangelogEntry{Port: 9000, Proto: "tcp", Event: "added", Timestamp: now.Add(2 * time.Second)})
	_ = store.Append(ChangelogEntry{Port: 8000, Proto: "tcp", Event: "added", Timestamp: now})

	all, err := store.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if all[0].Port != 8000 {
		t.Errorf("expected 8000 first, got %d", all[0].Port)
	}
}

func TestChangelogStore_FilterByPort(t *testing.T) {
	dir := t.TempDir()
	store := NewChangelogStore(dir)

	_ = store.Append(ChangelogEntry{Port: 80, Proto: "tcp", Event: "added"})
	_ = store.Append(ChangelogEntry{Port: 443, Proto: "tcp", Event: "added"})
	_ = store.Append(ChangelogEntry{Port: 80, Proto: "tcp", Event: "removed"})

	result, err := store.FilterByPort(80, "tcp")
	if err != nil {
		t.Fatalf("FilterByPort: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 entries for port 80, got %d", len(result))
	}
}

func TestChangelogStore_RecordDiff(t *testing.T) {
	dir := t.TempDir()
	store := NewChangelogStore(dir)

	diff := SnapshotDiff{
		Added:   []Listener{makeChangelogListener(3000, "tcp")},
		Removed: []Listener{makeChangelogListener(4000, "udp")},
	}
	if err := store.RecordDiff(diff); err != nil {
		t.Fatalf("RecordDiff: %v", err)
	}

	all, err := store.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 entries, got %d", len(all))
	}
	events := map[string]bool{}
	for _, e := range all {
		events[e.Event] = true
	}
	if !events["added"] || !events["removed"] {
		t.Errorf("expected both added and removed events, got: %v", events)
	}
}

func TestChangelogStore_SetsTimestampIfZero(t *testing.T) {
	dir := t.TempDir()
	store := NewChangelogStore(dir)

	before := time.Now().UTC().Add(-time.Second)
	_ = store.Append(ChangelogEntry{Port: 22, Proto: "tcp", Event: "added"})

	all, _ := store.All()
	if all[0].Timestamp.Before(before) {
		t.Errorf("expected timestamp to be set automatically")
	}
}

func TestChangelogStore_PersistsAcrossInstances(t *testing.T) {
	dir := t.TempDir()

	s1 := NewChangelogStore(dir)
	_ = s1.Append(ChangelogEntry{Port: 5432, Proto: "tcp", Event: "added"})

	s2 := NewChangelogStore(dir)
	all, err := s2.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(all) != 1 || all[0].Port != 5432 {
		t.Errorf("expected persisted entry, got: %+v", all)
	}
	_ = os.Remove(dir)
}
