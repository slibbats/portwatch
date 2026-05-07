package portscanner

import (
	"os"
	"testing"
)

func TestImpactStore_Set_And_Get(t *testing.T) {
	dir := t.TempDir()
	s, err := NewImpactStore(dir)
	if err != nil {
		t.Fatalf("NewImpactStore: %v", err)
	}
	if err := s.Set(443, "tcp", ImpactHigh, "public HTTPS"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	e, ok := s.Get(443, "tcp")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Level != ImpactHigh {
		t.Errorf("level = %q, want %q", e.Level, ImpactHigh)
	}
	if e.Rationale != "public HTTPS" {
		t.Errorf("rationale = %q, want %q", e.Rationale, "public HTTPS")
	}
}

func TestImpactStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewImpactStore(dir)
	_, ok := s.Get(9999, "tcp")
	if ok {
		t.Error("expected not found")
	}
}

func TestImpactStore_Remove(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewImpactStore(dir)
	_ = s.Set(80, "tcp", ImpactMedium, "")
	if err := s.Remove(80, "tcp"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	_, ok := s.Get(80, "tcp")
	if ok {
		t.Error("expected entry removed")
	}
}

func TestImpactStore_Remove_NotFound(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewImpactStore(dir)
	if err := s.Remove(1234, "udp"); err == nil {
		t.Error("expected error removing non-existent entry")
	}
}

func TestImpactStore_All_SortedByPort(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewImpactStore(dir)
	_ = s.Set(8080, "tcp", ImpactLow, "")
	_ = s.Set(22, "tcp", ImpactCritical, "")
	_ = s.Set(443, "tcp", ImpactHigh, "")
	all := s.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(all))
	}
	if all[0].Port != 22 || all[1].Port != 443 || all[2].Port != 8080 {
		t.Errorf("unexpected order: %v", all)
	}
}

func TestImpactStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	s1, _ := NewImpactStore(dir)
	_ = s1.Set(53, "udp", ImpactNone, "DNS")

	s2, err := NewImpactStore(dir)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	e, ok := s2.Get(53, "udp")
	if !ok {
		t.Fatal("expected persisted entry")
	}
	if e.Level != ImpactNone {
		t.Errorf("level = %q, want %q", e.Level, ImpactNone)
	}
}

func TestParseImpactLevel_Valid(t *testing.T) {
	for _, tc := range []string{"critical", "high", "medium", "low", "none"} {
		if _, err := ParseImpactLevel(tc); err != nil {
			t.Errorf("ParseImpactLevel(%q) unexpected error: %v", tc, err)
		}
	}
}

func TestParseImpactLevel_Invalid(t *testing.T) {
	if _, err := ParseImpactLevel("unknown"); err == nil {
		t.Error("expected error for unknown impact level")
	}
}

func TestImpactStore_MkdirError(t *testing.T) {
	// Use a file path as dir to force mkdir failure
	f, _ := os.CreateTemp("", "impact")
	f.Close()
	defer os.Remove(f.Name())
	_, err := NewImpactStore(f.Name())
	if err == nil {
		t.Error("expected error when dir is a file")
	}
}
