package portscanner

import (
	"os"
	"testing"
)

func TestRateLimitStore_Set_And_Get(t *testing.T) {
	dir := t.TempDir()
	s, err := NewRateLimitStore(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Set(8080, "tcp", 100); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	e, ok := s.Get(8080, "tcp")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.MaxPerHour != 100 {
		t.Errorf("expected 100, got %d", e.MaxPerHour)
	}
}

func TestRateLimitStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewRateLimitStore(dir)
	_, ok := s.Get(9999, "tcp")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestRateLimitStore_Remove(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewRateLimitStore(dir)
	_ = s.Set(443, "tcp", 50)
	if err := s.Remove(443, "tcp"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	_, ok := s.Get(443, "tcp")
	if ok {
		t.Fatal("expected entry to be removed")
	}
}

func TestRateLimitStore_Remove_NotFound(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewRateLimitStore(dir)
	if err := s.Remove(1234, "udp"); err == nil {
		t.Fatal("expected error for missing entry")
	}
}

func TestRateLimitStore_All_SortedByPort(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewRateLimitStore(dir)
	_ = s.Set(9000, "tcp", 10)
	_ = s.Set(80, "tcp", 200)
	_ = s.Set(443, "tcp", 150)
	all := s.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(all))
	}
	if all[0].Port != 80 || all[1].Port != 443 || all[2].Port != 9000 {
		t.Errorf("unexpected order: %v", all)
	}
}

func TestRateLimitStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	s1, _ := NewRateLimitStore(dir)
	_ = s1.Set(8080, "tcp", 60)

	s2, err := NewRateLimitStore(dir)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	e, ok := s2.Get(8080, "tcp")
	if !ok {
		t.Fatal("expected persisted entry")
	}
	if e.MaxPerHour != 60 {
		t.Errorf("expected 60, got %d", e.MaxPerHour)
	}
}

func TestNewRateLimitStore_MissingDir(t *testing.T) {
	_, err := NewRateLimitStore("/nonexistent/path/xyz")
	if err == nil {
		t.Fatal("expected error for bad dir")
	}
}

func init() {
	// ensure os is used
	_ = os.DevNull
}
