package portscanner

import (
	"os"
	"testing"
)

func TestTrustLevelStore_Set_And_Get(t *testing.T) {
	dir := t.TempDir()
	s, err := NewTrustLevelStore(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Set(8080, "tcp", TrustHigh); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	level, ok := s.Get(8080, "tcp")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if level != TrustHigh {
		t.Errorf("expected TrustHigh, got %v", level)
	}
}

func TestTrustLevelStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewTrustLevelStore(dir)
	_, ok := s.Get(9999, "tcp")
	if ok {
		t.Error("expected not found")
	}
}

func TestTrustLevelStore_Remove(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewTrustLevelStore(dir)
	_ = s.Set(443, "tcp", TrustVerified)
	if err := s.Remove(443, "tcp"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	_, ok := s.Get(443, "tcp")
	if ok {
		t.Error("expected entry to be removed")
	}
}

func TestTrustLevelStore_Remove_NotFound(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewTrustLevelStore(dir)
	if err := s.Remove(1234, "udp"); err == nil {
		t.Error("expected error for missing entry")
	}
}

func TestTrustLevelStore_All_SortedByPort(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewTrustLevelStore(dir)
	_ = s.Set(9000, "tcp", TrustLow)
	_ = s.Set(80, "tcp", TrustMedium)
	_ = s.Set(443, "tcp", TrustHigh)
	all := s.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(all))
	}
	if all[0].Port != 80 || all[1].Port != 443 || all[2].Port != 9000 {
		t.Errorf("unexpected order: %v", all)
	}
}

func TestTrustLevelStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	s1, _ := NewTrustLevelStore(dir)
	_ = s1.Set(22, "tcp", TrustVerified)
	_ = s1.Set(53, "udp", TrustMedium)

	s2, err := NewTrustLevelStore(dir)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	if l, ok := s2.Get(22, "tcp"); !ok || l != TrustVerified {
		t.Errorf("expected TrustVerified for port 22, got %v", l)
	}
	if l, ok := s2.Get(53, "udp"); !ok || l != TrustMedium {
		t.Errorf("expected TrustMedium for port 53, got %v", l)
	}
}

func TestParseTrustLevel_Valid(t *testing.T) {
	cases := []struct {
		input    string
		expected TrustLevel
	}{
		{"untrusted", TrustUntrusted},
		{"low", TrustLow},
		{"medium", TrustMedium},
		{"high", TrustHigh},
		{"verified", TrustVerified},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got, err := ParseTrustLevel(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

func TestParseTrustLevel_Invalid(t *testing.T) {
	_, err := ParseTrustLevel("bogus")
	if err == nil {
		t.Error("expected error for invalid trust level")
	}
}

func TestNewTrustLevelStore_CreatesDir(t *testing.T) {
	dir := t.TempDir()
	subdir := dir + "/nested/trustlevel"
	_, err := NewTrustLevelStore(subdir)
	if err != nil {
		t.Fatalf("expected dir creation, got: %v", err)
	}
	if _, err := os.Stat(subdir); os.IsNotExist(err) {
		t.Error("expected directory to be created")
	}
}
