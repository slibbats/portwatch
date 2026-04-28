package portscanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSeverityStore_Set_And_Get(t *testing.T) {
	dir := t.TempDir()
	s, err := NewSeverityStore(filepath.Join(dir, "severity.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Set(8080, "tcp", SeverityHigh); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	level, ok := s.Get(8080, "tcp")
	if !ok {
		t.Fatal("expected rule to exist")
	}
	if level != SeverityHigh {
		t.Errorf("expected high, got %s", level)
	}
}

func TestSeverityStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewSeverityStore(filepath.Join(dir, "severity.json"))
	level, ok := s.Get(9999, "tcp")
	if ok {
		t.Error("expected not found")
	}
	if level != SeverityLow {
		t.Errorf("expected default SeverityLow, got %s", level)
	}
}

func TestSeverityStore_Remove(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewSeverityStore(filepath.Join(dir, "severity.json"))
	_ = s.Set(443, "tcp", SeverityCritical)
	if err := s.Remove(443, "tcp"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	_, ok := s.Get(443, "tcp")
	if ok {
		t.Error("expected rule to be removed")
	}
}

func TestSeverityStore_All_ReturnsCopy(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewSeverityStore(filepath.Join(dir, "severity.json"))
	_ = s.Set(22, "tcp", SeverityMedium)
	_ = s.Set(80, "tcp", SeverityLow)
	all := s.All()
	if len(all) != 2 {
		t.Errorf("expected 2 rules, got %d", len(all))
	}
}

func TestSeverityStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "severity.json")
	s1, _ := NewSeverityStore(path)
	_ = s1.Set(3306, "tcp", SeverityCritical)

	s2, err := NewSeverityStore(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	level, ok := s2.Get(3306, "tcp")
	if !ok {
		t.Fatal("expected rule after reload")
	}
	if level != SeverityCritical {
		t.Errorf("expected critical, got %s", level)
	}
}

func TestSeverityStore_MissingFile_StartsEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent", "severity.json")
	// Parent dir doesn't exist yet — store should still init cleanly
	s, err := NewSeverityStore(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.All()) != 0 {
		t.Error("expected empty store")
	}
	// Ensure we can persist to a nested dir
	if err := s.Set(8443, "tcp", SeverityHigh); err != nil {
		t.Fatalf("Set to nested path failed: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file to be created: %v", err)
	}
}
