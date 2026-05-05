package portscanner

import (
	"os"
	"testing"
)

func TestScheduleStore_Set_And_Get(t *testing.T) {
	dir := t.TempDir()
	s, err := NewScheduleStore(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Set(8080, "tcp", "0 * * * *", "hourly check"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	e, ok := s.Get(8080, "tcp")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Cron != "0 * * * *" {
		t.Errorf("expected cron '0 * * * *', got %q", e.Cron)
	}
	if e.Label != "hourly check" {
		t.Errorf("expected label 'hourly check', got %q", e.Label)
	}
}

func TestScheduleStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewScheduleStore(dir)
	_, ok := s.Get(9999, "tcp")
	if ok {
		t.Error("expected not found")
	}
}

func TestScheduleStore_Remove(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewScheduleStore(dir)
	_ = s.Set(443, "tcp", "@daily", "")
	if err := s.Remove(443, "tcp"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	_, ok := s.Get(443, "tcp")
	if ok {
		t.Error("expected entry to be removed")
	}
}

func TestScheduleStore_Remove_NotFound(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewScheduleStore(dir)
	if err := s.Remove(1234, "udp"); err == nil {
		t.Error("expected error for missing entry")
	}
}

func TestScheduleStore_All_SortedByPort(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewScheduleStore(dir)
	_ = s.Set(9090, "tcp", "@hourly", "")
	_ = s.Set(80, "tcp", "@daily", "")
	_ = s.Set(443, "tcp", "@weekly", "")
	all := s.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(all))
	}
	if all[0].Port != 80 || all[1].Port != 443 || all[2].Port != 9090 {
		t.Errorf("unexpected order: %v", all)
	}
}

func TestScheduleStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	s1, _ := NewScheduleStore(dir)
	_ = s1.Set(22, "tcp", "*/5 * * * *", "ssh monitor")

	s2, err := NewScheduleStore(dir)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	e, ok := s2.Get(22, "tcp")
	if !ok {
		t.Fatal("expected entry after reload")
	}
	if e.Cron != "*/5 * * * *" {
		t.Errorf("cron mismatch after reload: %q", e.Cron)
	}
}

func TestScheduleStore_NewStore_NonExistentDir(t *testing.T) {
	dir := t.TempDir()
	_ = os.RemoveAll(dir)
	_ = os.MkdirAll(dir, 0755)
	s, err := NewScheduleStore(dir)
	if err != nil {
		t.Fatalf("unexpected error for empty dir: %v", err)
	}
	if len(s.All()) != 0 {
		t.Error("expected empty store")
	}
}
