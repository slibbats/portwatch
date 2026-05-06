package portscanner

import (
	"os"
	"testing"
)

func TestEscalationStore_Set_And_Get(t *testing.T) {
	dir := t.TempDir()
	s, err := NewEscalationStore(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Set(8080, "tcp", "ops@example.com", "#alerts", "high"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	p, ok := s.Get(8080, "tcp")
	if !ok {
		t.Fatal("expected policy to exist")
	}
	if p.Contact != "ops@example.com" {
		t.Errorf("expected contact ops@example.com, got %s", p.Contact)
	}
	if p.Channel != "#alerts" {
		t.Errorf("expected channel #alerts, got %s", p.Channel)
	}
	if p.MinLevel != "high" {
		t.Errorf("expected min_level high, got %s", p.MinLevel)
	}
}

func TestEscalationStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewEscalationStore(dir)
	_, ok := s.Get(9999, "tcp")
	if ok {
		t.Error("expected not found")
	}
}

func TestEscalationStore_Remove(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewEscalationStore(dir)
	_ = s.Set(443, "tcp", "sec@example.com", "#security", "critical")
	if err := s.Remove(443, "tcp"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	_, ok := s.Get(443, "tcp")
	if ok {
		t.Error("expected policy to be removed")
	}
}

func TestEscalationStore_Remove_NotFound(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewEscalationStore(dir)
	if err := s.Remove(1234, "tcp"); err == nil {
		t.Error("expected error removing non-existent policy")
	}
}

func TestEscalationStore_All_SortedByPort(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewEscalationStore(dir)
	_ = s.Set(9090, "tcp", "a@x.com", "#a", "low")
	_ = s.Set(80, "tcp", "b@x.com", "#b", "medium")
	_ = s.Set(443, "tcp", "c@x.com", "#c", "high")
	all := s.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 policies, got %d", len(all))
	}
	if all[0].Port != 80 || all[1].Port != 443 || all[2].Port != 9090 {
		t.Errorf("unexpected sort order: %v", all)
	}
}

func TestEscalationStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	s1, _ := NewEscalationStore(dir)
	_ = s1.Set(22, "tcp", "infra@example.com", "#infra", "critical")

	s2, err := NewEscalationStore(dir)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	p, ok := s2.Get(22, "tcp")
	if !ok {
		t.Fatal("expected policy after reload")
	}
	if p.Contact != "infra@example.com" {
		t.Errorf("expected infra@example.com, got %s", p.Contact)
	}
}

func TestParseEscalationPort_Valid(t *testing.T) {
	port, proto, err := parseEscalationPort("8080/tcp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != 8080 || proto != "tcp" {
		t.Errorf("got %d/%s", port, proto)
	}
}

func TestParseEscalationPort_Invalid(t *testing.T) {
	if _, _, err := parseEscalationPort("notaport/tcp"); err == nil {
		t.Error("expected error")
	}
	if _, _, err := parseEscalationPort("8080"); err == nil {
		t.Error("expected error for missing proto")
	}
}

func init() {
	_ = os.Getenv // suppress unused import
}
