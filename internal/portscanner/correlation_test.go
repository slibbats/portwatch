package portscanner

import (
	"os"
	"path/filepath"
	"testing"
)

func makeCorrelationListener(port int, proto string) Listener {
	return Listener{
		Port:     port,
		Protocol: proto,
		Address:  "0.0.0.0",
		PID:      1234,
		Process:  "testd",
	}
}

func TestCorrelationStore_Set_And_Get(t *testing.T) {
	dir := t.TempDir()
	store := NewCorrelationStore(dir)

	l := makeCorrelationListener(8080, "tcp")
	store.Set(l, "INC-001")

	got, ok := store.Get(l)
	if !ok {
		t.Fatal("expected correlation to exist")
	}
	if got != "INC-001" {
		t.Errorf("expected INC-001, got %s", got)
	}
}

func TestCorrelationStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	store := NewCorrelationStore(dir)

	l := makeCorrelationListener(9090, "tcp")
	_, ok := store.Get(l)
	if ok {
		t.Error("expected no correlation for unknown listener")
	}
}

func TestCorrelationStore_Remove(t *testing.T) {
	dir := t.TempDir()
	store := NewCorrelationStore(dir)

	l := makeCorrelationListener(443, "tcp")
	store.Set(l, "INC-002")
	store.Remove(l)

	_, ok := store.Get(l)
	if ok {
		t.Error("expected correlation to be removed")
	}
}

func TestCorrelationStore_All_SortedByPort(t *testing.T) {
	dir := t.TempDir()
	store := NewCorrelationStore(dir)

	store.Set(makeCorrelationListener(9000, "tcp"), "INC-003")
	store.Set(makeCorrelationListener(80, "tcp"), "INC-004")
	store.Set(makeCorrelationListener(443, "tcp"), "INC-005")

	all := store.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(all))
	}
	if all[0].Port > all[1].Port || all[1].Port > all[2].Port {
		t.Error("expected entries sorted by port")
	}
}

func TestCorrelationStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := NewCorrelationStore(dir)

	l := makeCorrelationListener(8443, "tcp")
	store.Set(l, "INC-006")

	if err := store.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	store2 := NewCorrelationStore(dir)
	if err := store2.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	got, ok := store2.Get(l)
	if !ok {
		t.Fatal("expected correlation after reload")
	}
	if got != "INC-006" {
		t.Errorf("expected INC-006, got %s", got)
	}
}

func TestCorrelationStore_Load_FileNotFound(t *testing.T) {
	dir := t.TempDir()
	store := NewCorrelationStore(filepath.Join(dir, "nonexistent"))
	if err := store.Load(); err != nil {
		if !os.IsNotExist(err) {
			t.Errorf("unexpected error: %v", err)
		}
	}
}
