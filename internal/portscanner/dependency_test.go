package portscanner

import (
	"os"
	"testing"
)

func TestDependencyStore_Set_And_Get(t *testing.T) {
	dir := t.TempDir()
	ds, err := NewDependencyStore(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	deps := []DependencyRef{{Port: 5432, Proto: "tcp", Note: "postgres"}}
	if err := ds.Set(8080, "tcp", deps); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	d, ok := ds.Get(8080, "tcp")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if len(d.DependsOn) != 1 || d.DependsOn[0].Port != 5432 {
		t.Errorf("unexpected deps: %+v", d.DependsOn)
	}
}

func TestDependencyStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	ds, _ := NewDependencyStore(dir)
	_, ok := ds.Get(9999, "tcp")
	if ok {
		t.Error("expected not found")
	}
}

func TestDependencyStore_Remove(t *testing.T) {
	dir := t.TempDir()
	ds, _ := NewDependencyStore(dir)
	_ = ds.Set(443, "tcp", []DependencyRef{{Port: 80, Proto: "tcp"}})
	if err := ds.Remove(443, "tcp"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	_, ok := ds.Get(443, "tcp")
	if ok {
		t.Error("expected entry to be removed")
	}
}

func TestDependencyStore_All_SortedByPort(t *testing.T) {
	dir := t.TempDir()
	ds, _ := NewDependencyStore(dir)
	_ = ds.Set(9090, "tcp", nil)
	_ = ds.Set(1234, "tcp", nil)
	_ = ds.Set(5000, "udp", nil)
	all := ds.All()
	if len(all) != 3 {
		t.Fatalf("expected 3, got %d", len(all))
	}
	if all[0].Port != 1234 || all[1].Port != 5000 || all[2].Port != 9090 {
		t.Errorf("unexpected order: %+v", all)
	}
}

func TestDependencyStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	ds, _ := NewDependencyStore(dir)
	_ = ds.Set(8443, "tcp", []DependencyRef{{Port: 8080, Proto: "tcp", Note: "upstream"}})

	ds2, err := NewDependencyStore(dir)
	if err != nil {
		t.Fatalf("reload error: %v", err)
	}
	d, ok := ds2.Get(8443, "tcp")
	if !ok {
		t.Fatal("expected entry after reload")
	}
	if d.DependsOn[0].Note != "upstream" {
		t.Errorf("note mismatch: %s", d.DependsOn[0].Note)
	}
}

func TestDependencyStore_Remove_NotFound(t *testing.T) {
	dir := t.TempDir()
	ds, _ := NewDependencyStore(dir)
	if err := ds.Remove(1111, "tcp"); err != nil {
		t.Errorf("expected no error removing non-existent entry, got: %v", err)
	}
	_ = os.Remove(dir) // ensure no panic
}
