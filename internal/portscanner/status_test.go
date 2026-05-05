package portscanner

import (
	"os"
	"testing"
)

func TestStatusStore_Set_And_Get(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStatusStore(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Set(8080, "tcp", StatusActive, "web server"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	e, ok := s.Get(8080, "tcp")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Status != StatusActive {
		t.Errorf("expected active, got %s", e.Status)
	}
	if e.Note != "web server" {
		t.Errorf("expected note 'web server', got %q", e.Note)
	}
}

func TestStatusStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStatusStore(dir)
	_, ok := s.Get(9999, "tcp")
	if ok {
		t.Error("expected not found")
	}
}

func TestStatusStore_Remove(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStatusStore(dir)
	_ = s.Set(443, "tcp", StatusActive, "")
	if err := s.Remove(443, "tcp"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	_, ok := s.Get(443, "tcp")
	if ok {
		t.Error("expected entry to be removed")
	}
}

func TestStatusStore_All_SortedByPort(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStatusStore(dir)
	_ = s.Set(9000, "tcp", StatusActive, "")
	_ = s.Set(80, "tcp", StatusInactive, "")
	_ = s.Set(443, "udp", StatusUnknown, "")
	all := s.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(all))
	}
	if all[0].Port != 80 || all[1].Port != 443 || all[2].Port != 9000 {
		t.Errorf("unexpected order: %v", all)
	}
}

func TestStatusStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	s1, _ := NewStatusStore(dir)
	_ = s1.Set(22, "tcp", StatusActive, "ssh")

	s2, err := NewStatusStore(dir)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	e, ok := s2.Get(22, "tcp")
	if !ok {
		t.Fatal("entry not persisted")
	}
	if e.Note != "ssh" {
		t.Errorf("expected note 'ssh', got %q", e.Note)
	}
}

func TestStatusStore_Set_UpdatesExisting(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStatusStore(dir)
	_ = s.Set(8080, "tcp", StatusActive, "old")
	_ = s.Set(8080, "tcp", StatusInactive, "new")
	e, _ := s.Get(8080, "tcp")
	if e.Status != StatusInactive {
		t.Errorf("expected inactive, got %s", e.Status)
	}
	if e.Note != "new" {
		t.Errorf("expected note 'new', got %q", e.Note)
	}
}

func TestNewStatusStore_MissingDir_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	subdir := dir + "/nested/store"
	s, err := NewStatusStore(subdir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Set(3000, "tcp", StatusActive, ""); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	if _, err := os.Stat(subdir + "/status.json"); err != nil {
		t.Errorf("expected file to be created: %v", err)
	}
}
