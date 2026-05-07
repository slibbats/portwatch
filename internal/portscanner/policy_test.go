package portscanner

import (
	"os"
	"testing"
)

func TestPolicyStore_Set_And_Get(t *testing.T) {
	dir := t.TempDir()
	ps, err := NewPolicyStore(dir)
	if err != nil {
		t.Fatalf("NewPolicyStore: %v", err)
	}
	if err := ps.Set(8080, "tcp", "allow", "internal service"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	p, ok := ps.Get(8080, "tcp")
	if !ok {
		t.Fatal("expected policy to exist")
	}
	if p.Action != "allow" || p.Reason != "internal service" {
		t.Errorf("unexpected policy: %+v", p)
	}
}

func TestPolicyStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	ps, _ := NewPolicyStore(dir)
	_, ok := ps.Get(9999, "tcp")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestPolicyStore_Remove(t *testing.T) {
	dir := t.TempDir()
	ps, _ := NewPolicyStore(dir)
	_ = ps.Set(443, "tcp", "deny", "blocked")
	if err := ps.Remove(443, "tcp"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	_, ok := ps.Get(443, "tcp")
	if ok {
		t.Fatal("expected policy to be removed")
	}
}

func TestPolicyStore_Remove_NotFound(t *testing.T) {
	dir := t.TempDir()
	ps, _ := NewPolicyStore(dir)
	if err := ps.Remove(1234, "udp"); err == nil {
		t.Fatal("expected error for missing policy")
	}
}

func TestPolicyStore_All_SortedByPort(t *testing.T) {
	dir := t.TempDir()
	ps, _ := NewPolicyStore(dir)
	_ = ps.Set(9000, "tcp", "allow", "")
	_ = ps.Set(80, "tcp", "deny", "")
	_ = ps.Set(443, "tcp", "allow", "")
	all := ps.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 policies, got %d", len(all))
	}
	if all[0].Port != 80 || all[1].Port != 443 || all[2].Port != 9000 {
		t.Errorf("unexpected order: %v", all)
	}
}

func TestPolicyStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	ps1, _ := NewPolicyStore(dir)
	_ = ps1.Set(22, "tcp", "allow", "ssh")

	ps2, err := NewPolicyStore(dir)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	p, ok := ps2.Get(22, "tcp")
	if !ok {
		t.Fatal("expected policy after reload")
	}
	if p.Reason != "ssh" {
		t.Errorf("unexpected reason: %s", p.Reason)
	}
}

func TestPolicyStore_MkdirOnCreate(t *testing.T) {
	dir := t.TempDir()
	subdir := dir + "/nested/policy"
	_, err := NewPolicyStore(subdir)
	if err != nil {
		t.Fatalf("expected dir creation: %v", err)
	}
	if _, err := os.Stat(subdir); os.IsNotExist(err) {
		t.Fatal("directory was not created")
	}
}
