package portscanner

import (
	"os"
	"testing"
)

func TestConfidenceStore_Set_And_Get(t *testing.T) {
	dir := t.TempDir()
	cs, err := NewConfidenceStore(dir)
	if err != nil {
		t.Fatalf("NewConfidenceStore: %v", err)
	}
	if err := cs.Set(8080, "tcp", ConfidenceHigh, "well-known service"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	e, ok := cs.Get(8080, "tcp")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Level != ConfidenceHigh {
		t.Errorf("level = %v, want high", e.Level)
	}
	if e.Rationale != "well-known service" {
		t.Errorf("rationale = %q, want 'well-known service'", e.Rationale)
	}
}

func TestConfidenceStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	cs, _ := NewConfidenceStore(dir)
	_, ok := cs.Get(9999, "tcp")
	if ok {
		t.Error("expected not found")
	}
}

func TestConfidenceStore_Remove(t *testing.T) {
	dir := t.TempDir()
	cs, _ := NewConfidenceStore(dir)
	_ = cs.Set(443, "tcp", ConfidenceMedium, "")
	if err := cs.Remove(443, "tcp"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	_, ok := cs.Get(443, "tcp")
	if ok {
		t.Error("expected entry to be removed")
	}
}

func TestConfidenceStore_Remove_NotFound(t *testing.T) {
	dir := t.TempDir()
	cs, _ := NewConfidenceStore(dir)
	if err := cs.Remove(1234, "udp"); err == nil {
		t.Error("expected error for missing entry")
	}
}

func TestConfidenceStore_All_SortedByPort(t *testing.T) {
	dir := t.TempDir()
	cs, _ := NewConfidenceStore(dir)
	_ = cs.Set(9000, "tcp", ConfidenceLow, "")
	_ = cs.Set(80, "tcp", ConfidenceHigh, "")
	_ = cs.Set(443, "tcp", ConfidenceMedium, "")
	all := cs.All()
	if len(all) != 3 {
		t.Fatalf("len = %d, want 3", len(all))
	}
	if all[0].Port != 80 || all[1].Port != 443 || all[2].Port != 9000 {
		t.Errorf("unexpected order: %v", all)
	}
}

func TestConfidenceStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	cs1, _ := NewConfidenceStore(dir)
	_ = cs1.Set(22, "tcp", ConfidenceHigh, "ssh")

	cs2, err := NewConfidenceStore(dir)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	e, ok := cs2.Get(22, "tcp")
	if !ok {
		t.Fatal("entry not persisted")
	}
	if e.Level != ConfidenceHigh || e.Rationale != "ssh" {
		t.Errorf("unexpected entry after reload: %+v", e)
	}
}

func TestParseConfidence_Valid(t *testing.T) {
	cases := []struct {
		input string
		want  ConfidenceLevel
	}{
		{"low", ConfidenceLow},
		{"medium", ConfidenceMedium},
		{"high", ConfidenceHigh},
	}
	for _, tc := range cases {
		got, err := ParseConfidence(tc.input)
		if err != nil {
			t.Errorf("ParseConfidence(%q): %v", tc.input, err)
		}
		if got != tc.want {
			t.Errorf("ParseConfidence(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestParseConfidence_Invalid(t *testing.T) {
	_, err := ParseConfidence("critical")
	if err == nil {
		t.Error("expected error for unknown confidence level")
	}
}

func TestNewConfidenceStore_CreatesDir(t *testing.T) {
	dir := t.TempDir()
	subdir := dir + "/nested/confidence"
	_, err := NewConfidenceStore(subdir)
	if err != nil {
		t.Fatalf("expected dir creation, got: %v", err)
	}
	if _, err := os.Stat(subdir); os.IsNotExist(err) {
		t.Error("directory was not created")
	}
}
