package portscanner

import (
	"os"
	"testing"
)

func TestCategoryStore_Set_And_Get(t *testing.T) {
	dir := t.TempDir()
	cs, err := NewCategoryStore(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := cs.Set(80, "tcp", "web"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	cat, ok := cs.Get(80, "tcp")
	if !ok {
		t.Fatal("expected category to exist")
	}
	if cat != "web" {
		t.Errorf("expected 'web', got %q", cat)
	}
}

func TestCategoryStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	cs, _ := NewCategoryStore(dir)
	_, ok := cs.Get(9999, "tcp")
	if ok {
		t.Error("expected not found")
	}
}

func TestCategoryStore_Remove(t *testing.T) {
	dir := t.TempDir()
	cs, _ := NewCategoryStore(dir)
	_ = cs.Set(443, "tcp", "secure-web")
	if err := cs.Remove(443, "tcp"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	_, ok := cs.Get(443, "tcp")
	if ok {
		t.Error("expected category to be removed")
	}
}

func TestCategoryStore_All_SortedByPort(t *testing.T) {
	dir := t.TempDir()
	cs, _ := NewCategoryStore(dir)
	_ = cs.Set(8080, "tcp", "alt-web")
	_ = cs.Set(22, "tcp", "ssh")
	_ = cs.Set(3306, "tcp", "database")
	all := cs.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(all))
	}
	if all[0].Port != 22 || all[1].Port != 3306 || all[2].Port != 8080 {
		t.Errorf("entries not sorted by port: %+v", all)
	}
}

func TestCategoryStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	cs1, _ := NewCategoryStore(dir)
	_ = cs1.Set(53, "udp", "dns")
	_ = cs1.Set(25, "tcp", "mail")

	cs2, err := NewCategoryStore(dir)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	cat, ok := cs2.Get(53, "udp")
	if !ok || cat != "dns" {
		t.Errorf("expected 'dns' for port 53/udp, got %q (ok=%v)", cat, ok)
	}
	cat, ok = cs2.Get(25, "tcp")
	if !ok || cat != "mail" {
		t.Errorf("expected 'mail' for port 25/tcp, got %q (ok=%v)", cat, ok)
	}
}

func TestCategoryStore_MissingDir_ReturnsNoError(t *testing.T) {
	dir := t.TempDir()
	nonExistent := dir + "/subdir"
	_ = os.MkdirAll(nonExistent, 0755)
	_, err := NewCategoryStore(nonExistent)
	if err != nil {
		t.Errorf("unexpected error for missing file: %v", err)
	}
}
