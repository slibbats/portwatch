package portscanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLabelStore_Set_And_Get(t *testing.T) {
	dir := t.TempDir()
	ls, err := NewLabelStore(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ls.Set(80, "tcp", "http")
	if got := ls.Get(80, "tcp"); got != "http" {
		t.Errorf("expected 'http', got %q", got)
	}
}

func TestLabelStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	ls, _ := NewLabelStore(dir)
	if got := ls.Get(9999, "tcp"); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestLabelStore_Remove(t *testing.T) {
	dir := t.TempDir()
	ls, _ := NewLabelStore(dir)
	ls.Set(443, "tcp", "https")
	ls.Remove(443, "tcp")
	if got := ls.Get(443, "tcp"); got != "" {
		t.Errorf("expected empty after remove, got %q", got)
	}
}

func TestLabelStore_All_ReturnsCopy(t *testing.T) {
	dir := t.TempDir()
	ls, _ := NewLabelStore(dir)
	ls.Set(22, "tcp", "ssh")
	ls.Set(53, "udp", "dns")
	all := ls.All()
	if len(all) != 2 {
		t.Errorf("expected 2 labels, got %d", len(all))
	}
	// Mutating the copy should not affect the store.
	delete(all, "22/tcp")
	if ls.Get(22, "tcp") != "ssh" {
		t.Error("store was mutated through All() copy")
	}
}

func TestLabelStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	ls, _ := NewLabelStore(dir)
	ls.Set(8080, "tcp", "dev-server")
	if err := ls.Save(); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	ls2, err := NewLabelStore(dir)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	if got := ls2.Get(8080, "tcp"); got != "dev-server" {
		t.Errorf("expected 'dev-server' after reload, got %q", got)
	}
}

func TestNewLabelStore_MissingFile(t *testing.T) {
	dir := t.TempDir()
	// Ensure no labels.json exists.
	os.Remove(filepath.Join(dir, "labels.json"))
	ls, err := NewLabelStore(dir)
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	if len(ls.All()) != 0 {
		t.Error("expected empty store for missing file")
	}
}
