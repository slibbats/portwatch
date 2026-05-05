package portscanner

import (
	"os"
	"testing"
	"time"
)

func TestExpiryStore_Set_And_Get(t *testing.T) {
	dir := t.TempDir()
	s, err := NewExpiryStore(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expiry := time.Now().Add(24 * time.Hour)
	if err := s.Set(8080, "tcp", expiry, "temporary dev server"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	e, ok := s.Get(8080, "tcp")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Port != 8080 || e.Proto != "tcp" {
		t.Errorf("unexpected entry: %+v", e)
	}
	if e.Note != "temporary dev server" {
		t.Errorf("unexpected note: %q", e.Note)
	}
}

func TestExpiryStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewExpiryStore(dir)
	_, ok := s.Get(9999, "tcp")
	if ok {
		t.Fatal("expected no entry")
	}
}

func TestExpiryStore_Remove(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewExpiryStore(dir)
	_ = s.Set(443, "tcp", time.Now().Add(time.Hour), "")

	if err := s.Remove(443, "tcp"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	_, ok := s.Get(443, "tcp")
	if ok {
		t.Fatal("expected entry to be removed")
	}
}

func TestExpiryStore_Remove_NotFound(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewExpiryStore(dir)
	if err := s.Remove(1234, "udp"); err == nil {
		t.Fatal("expected error for missing entry")
	}
}

func TestExpiryStore_Expired(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewExpiryStore(dir)

	_ = s.Set(80, "tcp", time.Now().Add(-time.Minute), "already expired")
	_ = s.Set(443, "tcp", time.Now().Add(time.Hour), "still valid")

	expired := s.Expired()
	if len(expired) != 1 {
		t.Fatalf("expected 1 expired entry, got %d", len(expired))
	}
	if expired[0].Port != 80 {
		t.Errorf("expected port 80, got %d", expired[0].Port)
	}
}

func TestExpiryStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	s1, _ := NewExpiryStore(dir)
	expiry := time.Now().Add(48 * time.Hour).Truncate(time.Second)
	_ = s1.Set(3000, "tcp", expiry, "roundtrip test")

	s2, err := NewExpiryStore(dir)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	e, ok := s2.Get(3000, "tcp")
	if !ok {
		t.Fatal("expected entry after reload")
	}
	if !e.ExpiresAt.Equal(expiry) {
		t.Errorf("expiry mismatch: got %v, want %v", e.ExpiresAt, expiry)
	}
}

func TestExpiryStore_All_SortedByPort(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewExpiryStore(dir)
	_ = s.Set(9090, "tcp", time.Now().Add(time.Hour), "")
	_ = s.Set(1080, "tcp", time.Now().Add(time.Hour), "")
	_ = s.Set(5432, "tcp", time.Now().Add(time.Hour), "")

	all := s.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(all))
	}
	if all[0].Port != 1080 || all[1].Port != 5432 || all[2].Port != 9090 {
		t.Errorf("unexpected order: %v", all)
	}
}

func TestNewExpiryStore_MissingFileIsOk(t *testing.T) {
	dir := t.TempDir()
	_ = os.Remove(dir + "/expiry.json")
	_, err := NewExpiryStore(dir)
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
}
