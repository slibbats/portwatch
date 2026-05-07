package portscanner

import (
	"os"
	"testing"
)

func TestHostmapStore_Set_And_Get(t *testing.T) {
	dir := t.TempDir()
	s, err := NewHostmapStore(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Set(8080, "tcp", "api.internal"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	host, ok := s.Get(8080, "tcp")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if host != "api.internal" {
		t.Errorf("expected api.internal, got %s", host)
	}
}

func TestHostmapStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewHostmapStore(dir)
	_, ok := s.Get(9999, "tcp")
	if ok {
		t.Error("expected not found")
	}
}

func TestHostmapStore_Remove(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewHostmapStore(dir)
	_ = s.Set(443, "tcp", "secure.host")
	if err := s.Remove(443, "tcp"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	_, ok := s.Get(443, "tcp")
	if ok {
		t.Error("expected entry to be removed")
	}
}

func TestHostmapStore_Remove_NotFound(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewHostmapStore(dir)
	if err := s.Remove(1234, "tcp"); err == nil {
		t.Error("expected error for missing entry")
	}
}

func TestHostmapStore_All_SortedByPort(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewHostmapStore(dir)
	_ = s.Set(9000, "tcp", "z.host")
	_ = s.Set(80, "tcp", "a.host")
	_ = s.Set(443, "udp", "b.host")
	all := s.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(all))
	}
	if all[0].Port != 80 {
		t.Errorf("expected first port 80, got %d", all[0].Port)
	}
	if all[2].Port != 9000 {
		t.Errorf("expected last port 9000, got %d", all[2].Port)
	}
}

func TestHostmapStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	s1, _ := NewHostmapStore(dir)
	_ = s1.Set(8443, "tcp", "vault.internal")

	s2, err := NewHostmapStore(dir)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	host, ok := s2.Get(8443, "tcp")
	if !ok {
		t.Fatal("expected reloaded entry")
	}
	if host != "vault.internal" {
		t.Errorf("expected vault.internal, got %s", host)
	}
	_ = os.RemoveAll(dir)
}
