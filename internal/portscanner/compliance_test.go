package portscanner

import (
	"os"
	"testing"
)

func TestComplianceStore_Set_And_Get(t *testing.T) {
	dir := t.TempDir()
	cs, err := NewComplianceStore(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := cs.Set(8080, "tcp", CompliancePass, "PCI-DSS", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}
	e, ok := cs.Get(8080, "tcp")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Status != CompliancePass {
		t.Errorf("expected pass, got %s", e.Status)
	}
	if e.Policy != "PCI-DSS" {
		t.Errorf("expected PCI-DSS, got %s", e.Policy)
	}
}

func TestComplianceStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	cs, _ := NewComplianceStore(dir)
	_, ok := cs.Get(9999, "tcp")
	if ok {
		t.Error("expected not found")
	}
}

func TestComplianceStore_Remove(t *testing.T) {
	dir := t.TempDir()
	cs, _ := NewComplianceStore(dir)
	_ = cs.Set(443, "tcp", ComplianceFail, "HIPAA", "missing TLS")
	if err := cs.Remove(443, "tcp"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	_, ok := cs.Get(443, "tcp")
	if ok {
		t.Error("expected entry to be removed")
	}
}

func TestComplianceStore_Remove_NotFound(t *testing.T) {
	dir := t.TempDir()
	cs, _ := NewComplianceStore(dir)
	if err := cs.Remove(1234, "udp"); err == nil {
		t.Error("expected error for missing entry")
	}
}

func TestComplianceStore_All_SortedByPort(t *testing.T) {
	dir := t.TempDir()
	cs, _ := NewComplianceStore(dir)
	_ = cs.Set(9090, "tcp", ComplianceWarning, "SOC2", "")
	_ = cs.Set(80, "tcp", CompliancePass, "SOC2", "")
	_ = cs.Set(443, "tcp", ComplianceFail, "SOC2", "expired cert")
	all := cs.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(all))
	}
	if all[0].Port != 80 || all[1].Port != 443 || all[2].Port != 9090 {
		t.Errorf("unexpected order: %v", all)
	}
}

func TestComplianceStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	cs, _ := NewComplianceStore(dir)
	_ = cs.Set(22, "tcp", ComplianceFail, "CIS", "SSH exposed")

	cs2, err := NewComplianceStore(dir)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	e, ok := cs2.Get(22, "tcp")
	if !ok {
		t.Fatal("expected entry after reload")
	}
	if e.Reason != "SSH exposed" {
		t.Errorf("expected reason 'SSH exposed', got %q", e.Reason)
	}
}

func TestParseComplianceStatus_Valid(t *testing.T) {
	for _, s := range []string{"pass", "fail", "warning", "unknown"} {
		if _, err := ParseComplianceStatus(s); err != nil {
			t.Errorf("expected valid status for %q: %v", s, err)
		}
	}
}

func TestParseComplianceStatus_Invalid(t *testing.T) {
	if _, err := ParseComplianceStatus("bogus"); err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestNewComplianceStore_BadDir(t *testing.T) {
	// Use a file as the dir to force mkdir failure
	f, _ := os.CreateTemp("", "compliance")
	defer os.Remove(f.Name())
	f.Close()
	if _, err := NewComplianceStore(f.Name()); err == nil {
		t.Error("expected error for bad dir")
	}
}
