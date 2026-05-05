package portscanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPriorityStore_Set_And_Get(t *testing.T) {
	dir := t.TempDir()
	ps, err := NewPriorityStore(dir)
	if err != nil {
		t.Fatalf("NewPriorityStore: %v", err)
	}
	if err := ps.Set(8080, "tcp", PriorityHigh); err != nil {
		t.Fatalf("Set: %v", err)
	}
	p, ok := ps.Get(8080, "tcp")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if p != PriorityHigh {
		t.Errorf("got %v, want %v", p, PriorityHigh)
	}
}

func TestPriorityStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	ps, _ := NewPriorityStore(dir)
	_, ok := ps.Get(9999, "tcp")
	if ok {
		t.Error("expected not found")
	}
}

func TestPriorityStore_Remove(t *testing.T) {
	dir := t.TempDir()
	ps, _ := NewPriorityStore(dir)
	_ = ps.Set(443, "tcp", PriorityCritical)
	if err := ps.Remove(443, "tcp"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	_, ok := ps.Get(443, "tcp")
	if ok {
		t.Error("expected entry to be removed")
	}
}

func TestPriorityStore_Remove_NotFound(t *testing.T) {
	dir := t.TempDir()
	ps, _ := NewPriorityStore(dir)
	if err := ps.Remove(1234, "udp"); err == nil {
		t.Error("expected error for missing entry")
	}
}

func TestPriorityStore_All_SortedByPort(t *testing.T) {
	dir := t.TempDir()
	ps, _ := NewPriorityStore(dir)
	_ = ps.Set(9000, "tcp", PriorityLow)
	_ = ps.Set(80, "tcp", PriorityMedium)
	_ = ps.Set(443, "tcp", PriorityCritical)
	all := ps.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(all))
	}
	if all[0].Port != 80 || all[1].Port != 443 || all[2].Port != 9000 {
		t.Errorf("unexpected order: %v", all)
	}
}

func TestPriorityStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	ps1, _ := NewPriorityStore(dir)
	_ = ps1.Set(22, "tcp", PriorityHigh)
	_ = ps1.Set(53, "udp", PriorityMedium)

	ps2, err := NewPriorityStore(dir)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	p, ok := ps2.Get(22, "tcp")
	if !ok || p != PriorityHigh {
		t.Errorf("port 22: got %v ok=%v, want high", p, ok)
	}
	p, ok = ps2.Get(53, "udp")
	if !ok || p != PriorityMedium {
		t.Errorf("port 53: got %v ok=%v, want medium", p, ok)
	}
}

func TestParsePriority_Valid(t *testing.T) {
	cases := []struct {
		input string
		want  Priority
	}{
		{"low", PriorityLow},
		{"medium", PriorityMedium},
		{"high", PriorityHigh},
		{"critical", PriorityCritical},
	}
	for _, c := range cases {
		p, err := ParsePriority(c.input)
		if err != nil {
			t.Errorf("%q: unexpected error: %v", c.input, err)
		}
		if p != c.want {
			t.Errorf("%q: got %v, want %v", c.input, p, c.want)
		}
	}
}

func TestParsePriority_Invalid(t *testing.T) {
	_, err := ParsePriority("extreme")
	if err == nil {
		t.Error("expected error for unknown priority")
	}
}

func TestNewPriorityStore_MissingDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nonexistent")
	ps, err := NewPriorityStore(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := ps.Set(80, "tcp", PriorityLow); err != nil {
		t.Errorf("Set in new dir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "priorities.json")); err != nil {
		t.Errorf("file not created: %v", err)
	}
}
