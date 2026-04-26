package portscanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGroupStore_Add_And_Get(t *testing.T) {
	gs, err := NewGroupStore(filepath.Join(t.TempDir(), "groups.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	gs.Add("web", 80)
	gs.Add("web", 443)
	g, ok := gs.Get("web")
	if !ok {
		t.Fatal("expected group 'web' to exist")
	}
	if len(g.Ports) != 2 {
		t.Errorf("expected 2 ports, got %d", len(g.Ports))
	}
}

func TestGroupStore_Add_NoDuplicates(t *testing.T) {
	gs, _ := NewGroupStore(filepath.Join(t.TempDir(), "groups.json"))
	gs.Add("db", 5432)
	gs.Add("db", 5432)
	g, _ := gs.Get("db")
	if len(g.Ports) != 1 {
		t.Errorf("expected 1 port, got %d", len(g.Ports))
	}
}

func TestGroupStore_Remove(t *testing.T) {
	gs, _ := NewGroupStore(filepath.Join(t.TempDir(), "groups.json"))
	gs.Add("web", 80)
	gs.Add("web", 443)
	gs.Remove("web", 80)
	g, ok := gs.Get("web")
	if !ok {
		t.Fatal("group should still exist after partial remove")
	}
	if len(g.Ports) != 1 || g.Ports[0] != 443 {
		t.Errorf("expected only port 443, got %v", g.Ports)
	}
}

func TestGroupStore_Remove_DeletesEmptyGroup(t *testing.T) {
	gs, _ := NewGroupStore(filepath.Join(t.TempDir(), "groups.json"))
	gs.Add("solo", 9090)
	gs.Remove("solo", 9090)
	_, ok := gs.Get("solo")
	if ok {
		t.Error("expected group to be deleted when empty")
	}
}

func TestGroupStore_All_SortedByName(t *testing.T) {
	gs, _ := NewGroupStore(filepath.Join(t.TempDir(), "groups.json"))
	gs.Add("zebra", 1)
	gs.Add("alpha", 2)
	gs.Add("middle", 3)
	all := gs.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(all))
	}
	if all[0].Name != "alpha" || all[1].Name != "middle" || all[2].Name != "zebra" {
		t.Errorf("groups not sorted: %v", all)
	}
}

func TestGroupStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "groups.json")
	gs, _ := NewGroupStore(path)
	gs.Add("web", 80)
	gs.Add("web", 443)
	gs.Add("db", 5432)
	if err := gs.Save(); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	gs2, err := NewGroupStore(path)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	g, ok := gs2.Get("web")
	if !ok || len(g.Ports) != 2 {
		t.Errorf("expected web group with 2 ports after reload")
	}
}

func TestGroupStore_Get_NotFound(t *testing.T) {
	gs, _ := NewGroupStore(filepath.Join(t.TempDir(), "groups.json"))
	_, ok := gs.Get("nonexistent")
	if ok {
		t.Error("expected not found")
	}
}

func TestNewGroupStore_MissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "no_such_file.json")
	gs, err := NewGroupStore(path)
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if gs == nil {
		t.Fatal("expected non-nil store")
	}
	_ = os.Remove(path)
}
