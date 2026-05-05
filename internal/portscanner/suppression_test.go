package portscanner

import (
	"os"
	"testing"
	"time"
)

func TestSuppressionStore_Set_And_Get(t *testing.T) {
	dir := t.TempDir()
	s, err := NewSuppressionStore(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	until := time.Now().Add(1 * time.Hour)
	if err := s.Set(8080, "tcp", "maintenance", until); err != nil {
		t.Fatalf("Set: %v", err)
	}
	e, ok := s.Get(8080, "tcp")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Reason != "maintenance" {
		t.Errorf("reason = %q, want %q", e.Reason, "maintenance")
	}
}

func TestSuppressionStore_IsSuppressed_Active(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewSuppressionStore(dir)
	_ = s.Set(443, "tcp", "planned", time.Now().Add(time.Hour))
	if !s.IsSuppressed(443, "tcp") {
		t.Error("expected port to be suppressed")
	}
}

func TestSuppressionStore_IsSuppressed_Expired(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewSuppressionStore(dir)
	_ = s.Set(443, "tcp", "old", time.Now().Add(-1*time.Second))
	if s.IsSuppressed(443, "tcp") {
		t.Error("expected expired suppression to be inactive")
	}
}

func TestSuppressionStore_Remove(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewSuppressionStore(dir)
	_ = s.Set(22, "tcp", "test", time.Now().Add(time.Hour))
	if err := s.Remove(22, "tcp"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if _, ok := s.Get(22, "tcp"); ok {
		t.Error("expected entry to be removed")
	}
}

func TestSuppressionStore_Remove_NotFound(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewSuppressionStore(dir)
	if err := s.Remove(9999, "tcp"); err == nil {
		t.Error("expected error for missing entry")
	}
}

func TestSuppressionStore_PruneExpired(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewSuppressionStore(dir)
	_ = s.Set(1000, "tcp", "expired", time.Now().Add(-time.Second))
	_ = s.Set(2000, "tcp", "active", time.Now().Add(time.Hour))
	if err := s.PruneExpired(); err != nil {
		t.Fatalf("PruneExpired: %v", err)
	}
	if _, ok := s.Get(1000, "tcp"); ok {
		t.Error("expected expired entry to be pruned")
	}
	if _, ok := s.Get(2000, "tcp"); !ok {
		t.Error("expected active entry to remain")
	}
}

func TestSuppressionStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewSuppressionStore(dir)
	until := time.Now().Add(2 * time.Hour).Truncate(time.Second)
	_ = s.Set(3306, "tcp", "db-maint", until)

	s2, err := NewSuppressionStore(dir)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	e, ok := s2.Get(3306, "tcp")
	if !ok {
		t.Fatal("expected entry after reload")
	}
	if e.Reason != "db-maint" {
		t.Errorf("reason = %q, want %q", e.Reason, "db-maint")
	}
}

func TestSuppressionStore_All_SortedByPort(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewSuppressionStore(dir)
	_ = s.Set(9000, "tcp", "", time.Now().Add(time.Hour))
	_ = s.Set(80, "tcp", "", time.Now().Add(time.Hour))
	_ = s.Set(443, "tcp", "", time.Now().Add(time.Hour))
	all := s.All()
	if len(all) != 3 {
		t.Fatalf("len = %d, want 3", len(all))
	}
	if all[0].Port != 80 || all[1].Port != 443 || all[2].Port != 9000 {
		t.Errorf("unexpected order: %v", all)
	}
}

func TestNewSuppressionStore_MissingDir(t *testing.T) {
	dir := t.TempDir()
	_ = os.RemoveAll(dir)
	s, err := NewSuppressionStore(dir)
	if err != nil {
		t.Fatalf("unexpected error for missing dir: %v", err)
	}
	if s == nil {
		t.Error("expected non-nil store")
	}
}
