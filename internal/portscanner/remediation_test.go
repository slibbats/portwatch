package portscanner

import (
	"os"
	"testing"
)

func TestRemediationStore_Set_And_Get(t *testing.T) {
	dir := t.TempDir()
	s, err := NewRemediationStore(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := RemediationAction{Port: 8080, Proto: "tcp", Action: "block", Script: "/usr/local/bin/block.sh"}
	if err := s.Set(r); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	got, ok := s.Get(8080, "tcp")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if got.Action != "block" || got.Script != "/usr/local/bin/block.sh" {
		t.Errorf("unexpected value: %+v", got)
	}
}

func TestRemediationStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewRemediationStore(dir)
	_, ok := s.Get(9999, "tcp")
	if ok {
		t.Error("expected not found")
	}
}

func TestRemediationStore_Remove(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewRemediationStore(dir)
	s.Set(RemediationAction{Port: 22, Proto: "tcp", Action: "alert"})
	if err := s.Remove(22, "tcp"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	_, ok := s.Get(22, "tcp")
	if ok {
		t.Error("expected entry to be removed")
	}
}

func TestRemediationStore_Remove_NotFound(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewRemediationStore(dir)
	if err := s.Remove(1234, "udp"); err == nil {
		t.Error("expected error for missing entry")
	}
}

func TestRemediationStore_All_SortedByPort(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewRemediationStore(dir)
	s.Set(RemediationAction{Port: 443, Proto: "tcp", Action: "log"})
	s.Set(RemediationAction{Port: 80, Proto: "tcp", Action: "block"})
	s.Set(RemediationAction{Port: 22, Proto: "tcp", Action: "alert"})
	all := s.All()
	if len(all) != 3 {
		t.Fatalf("expected 3, got %d", len(all))
	}
	if all[0].Port != 22 || all[1].Port != 80 || all[2].Port != 443 {
		t.Errorf("unexpected order: %v", all)
	}
}

func TestRemediationStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	s1, _ := NewRemediationStore(dir)
	s1.Set(RemediationAction{Port: 3306, Proto: "tcp", Action: "isolate", RunAs: "root"})

	s2, err := NewRemediationStore(dir)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	got, ok := s2.Get(3306, "tcp")
	if !ok {
		t.Fatal("expected persisted entry")
	}
	if got.RunAs != "root" {
		t.Errorf("expected RunAs=root, got %q", got.RunAs)
	}
}

func TestRemediationStore_NewDir_CreatesDir(t *testing.T) {
	base := t.TempDir()
	dir := base + "/nested/remediation"
	_, err := NewRemediationStore(dir)
	if err != nil {
		t.Fatalf("expected dir creation: %v", err)
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("directory was not created")
	}
}
