package portscanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNoteStore_Set_And_Get(t *testing.T) {
	dir := t.TempDir()
	ns, err := NewNoteStore(filepath.Join(dir, "notes.json"))
	if err != nil {
		t.Fatalf("NewNoteStore: %v", err)
	}
	ns.Set(8080, "tcp", "web server")
	n, ok := ns.Get(8080, "tcp")
	if !ok {
		t.Fatal("expected note to exist")
	}
	if n.Text != "web server" {
		t.Errorf("got text %q, want %q", n.Text, "web server")
	}
}

func TestNoteStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	ns, _ := NewNoteStore(filepath.Join(dir, "notes.json"))
	_, ok := ns.Get(9999, "tcp")
	if ok {
		t.Error("expected note not to exist")
	}
}

func TestNoteStore_Set_UpdatesExisting(t *testing.T) {
	dir := t.TempDir()
	ns, _ := NewNoteStore(filepath.Join(dir, "notes.json"))
	ns.Set(443, "tcp", "https")
	ns.Set(443, "tcp", "https updated")
	n, ok := ns.Get(443, "tcp")
	if !ok {
		t.Fatal("expected note")
	}
	if n.Text != "https updated" {
		t.Errorf("expected updated text, got %q", n.Text)
	}
}

func TestNoteStore_Remove(t *testing.T) {
	dir := t.TempDir()
	ns, _ := NewNoteStore(filepath.Join(dir, "notes.json"))
	ns.Set(22, "tcp", "ssh")
	removed := ns.Remove(22, "tcp")
	if !removed {
		t.Error("expected Remove to return true")
	}
	_, ok := ns.Get(22, "tcp")
	if ok {
		t.Error("expected note to be removed")
	}
}

func TestNoteStore_Remove_NotFound(t *testing.T) {
	dir := t.TempDir()
	ns, _ := NewNoteStore(filepath.Join(dir, "notes.json"))
	if ns.Remove(1234, "udp") {
		t.Error("expected Remove to return false for missing note")
	}
}

func TestNoteStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "notes.json")
	ns, _ := NewNoteStore(path)
	ns.Set(80, "tcp", "http")
	ns.Set(53, "udp", "dns")
	if err := ns.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	ns2, err := NewNoteStore(path)
	if err != nil {
		t.Fatalf("NewNoteStore reload: %v", err)
	}
	if n, ok := ns2.Get(80, "tcp"); !ok || n.Text != "http" {
		t.Errorf("expected http note, got %v %v", n, ok)
	}
	if n, ok := ns2.Get(53, "udp"); !ok || n.Text != "dns" {
		t.Errorf("expected dns note, got %v %v", n, ok)
	}
}

func TestNoteStore_All_ReturnsCopy(t *testing.T) {
	dir := t.TempDir()
	ns, _ := NewNoteStore(filepath.Join(dir, "notes.json"))
	ns.Set(8080, "tcp", "dev")
	ns.Set(9090, "tcp", "metrics")
	all := ns.All()
	if len(all) != 2 {
		t.Errorf("expected 2 notes, got %d", len(all))
	}
	all[0].Text = "mutated"
	if n, _ := ns.Get(all[0].Port, all[0].Protocol); n.Text == "mutated" {
		t.Error("All() should return copies, not references")
	}
}

func TestNewNoteStore_MissingFileIsOK(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent", "notes.json")
	_, err := NewNoteStore(path)
	if err != nil && !os.IsNotExist(err) {
		t.Errorf("unexpected error: %v", err)
	}
}
