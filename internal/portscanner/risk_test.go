package portscanner

import (
	"os"
	"testing"
)

func TestRiskStore_Set_And_Get(t *testing.T) {
	dir := t.TempDir()
	rs, err := NewRiskStore(dir)
	if err != nil {
		t.Fatalf("NewRiskStore: %v", err)
	}
	if err := rs.Set(8080, "tcp", RiskHigh); err != nil {
		t.Fatalf("Set: %v", err)
	}
	level, ok := rs.Get(8080, "tcp")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if level != RiskHigh {
		t.Errorf("expected RiskHigh, got %v", level)
	}
}

func TestRiskStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	rs, _ := NewRiskStore(dir)
	_, ok := rs.Get(9999, "tcp")
	if ok {
		t.Error("expected not found")
	}
}

func TestRiskStore_Remove(t *testing.T) {
	dir := t.TempDir()
	rs, _ := NewRiskStore(dir)
	_ = rs.Set(443, "tcp", RiskLow)
	if err := rs.Remove(443, "tcp"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	_, ok := rs.Get(443, "tcp")
	if ok {
		t.Error("expected entry to be removed")
	}
}

func TestRiskStore_Remove_NotFound(t *testing.T) {
	dir := t.TempDir()
	rs, _ := NewRiskStore(dir)
	if err := rs.Remove(1234, "udp"); err == nil {
		t.Error("expected error for missing entry")
	}
}

func TestRiskStore_All_SortedByPort(t *testing.T) {
	dir := t.TempDir()
	rs, _ := NewRiskStore(dir)
	_ = rs.Set(9090, "tcp", RiskMedium)
	_ = rs.Set(80, "tcp", RiskLow)
	_ = rs.Set(443, "tcp", RiskCritical)
	all := rs.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(all))
	}
	if all[0].Port != 80 || all[1].Port != 443 || all[2].Port != 9090 {
		t.Errorf("unexpected order: %v", all)
	}
}

func TestRiskStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	rs, _ := NewRiskStore(dir)
	_ = rs.Set(22, "tcp", RiskCritical)

	rs2, err := NewRiskStore(dir)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	level, ok := rs2.Get(22, "tcp")
	if !ok {
		t.Fatal("expected persisted entry")
	}
	if level != RiskCritical {
		t.Errorf("expected RiskCritical, got %v", level)
	}
}

func TestParseRiskLevel_Valid(t *testing.T) {
	cases := []struct {
		input    string
		expected RiskLevel
	}{
		{"low", RiskLow},
		{"medium", RiskMedium},
		{"high", RiskHigh},
		{"critical", RiskCritical},
	}
	for _, c := range cases {
		got, err := ParseRiskLevel(c.input)
		if err != nil {
			t.Errorf("ParseRiskLevel(%q): unexpected error: %v", c.input, err)
		}
		if got != c.expected {
			t.Errorf("ParseRiskLevel(%q): got %v, want %v", c.input, got, c.expected)
		}
	}
}

func TestParseRiskLevel_Invalid(t *testing.T) {
	_, err := ParseRiskLevel("extreme")
	if err == nil {
		t.Error("expected error for unknown risk level")
	}
}

func TestRiskStore_NewStore_MissingDir(t *testing.T) {
	dir := t.TempDir()
	_ = os.RemoveAll(dir)
	rs, err := NewRiskStore(dir)
	if err != nil {
		t.Fatalf("expected no error for missing dir, got: %v", err)
	}
	if err := rs.Set(80, "tcp", RiskLow); err != nil {
		t.Errorf("Set after missing dir: %v", err)
	}
}
