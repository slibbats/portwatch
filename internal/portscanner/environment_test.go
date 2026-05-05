package portscanner

import (
	"os"
	"testing"
)

func TestEnvironmentStore_Set_And_Get(t *testing.T) {
	dir := t.TempDir()
	s, err := NewEnvironmentStore(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Set(8080, "tcp", "production"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	e, ok := s.Get(8080, "tcp")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Environment != "production" {
		t.Errorf("expected 'production', got %q", e.Environment)
	}
}

func TestEnvironmentStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewEnvironmentStore(dir)
	_, ok := s.Get(9999, "tcp")
	if ok {
		t.Error("expected not found")
	}
}

func TestEnvironmentStore_Remove(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewEnvironmentStore(dir)
	_ = s.Set(443, "tcp", "staging")
	if err := s.Remove(443, "tcp"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	_, ok := s.Get(443, "tcp")
	if ok {
		t.Error("expected entry to be removed")
	}
}

func TestEnvironmentStore_All_SortedByPort(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewEnvironmentStore(dir)
	_ = s.Set(9000, "tcp", "dev")
	_ = s.Set(80, "tcp", "production")
	_ = s.Set(443, "udp", "staging")
	all := s.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(all))
	}
	if all[0].Port != 80 || all[1].Port != 443 || all[2].Port != 9000 {
		t.Errorf("unexpected order: %v", all)
	}
}

func TestEnvironmentStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	s1, _ := NewEnvironmentStore(dir)
	_ = s1.Set(5432, "tcp", "production")

	s2, err := NewEnvironmentStore(dir)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	e, ok := s2.Get(5432, "tcp")
	if !ok {
		t.Fatal("expected entry after reload")
	}
	if e.Environment != "production" {
		t.Errorf("expected 'production', got %q", e.Environment)
	}
}

func TestEnvironmentStore_MissingDir_ReturnsError(t *testing.T) {
	_, err := NewEnvironmentStore("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("expected error for non-existent directory")
	}
}

func init() {
	// Ensure the missing-dir test does not create files accidentally.
	_ = os.MkdirAll
}
