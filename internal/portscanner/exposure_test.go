package portscanner

import (
	"testing"
)

func TestExposureStore_Set_And_Get(t *testing.T) {
	dir := t.TempDir()
	s, err := NewExposureStore(dir)
	if err != nil {
		t.Fatalf("NewExposureStore: %v", err)
	}
	if err := s.Set(8080, "tcp", ExposurePublic); err != nil {
		t.Fatalf("Set: %v", err)
	}
	level, ok := s.Get(8080, "tcp")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if level != ExposurePublic {
		t.Errorf("got %q, want %q", level, ExposurePublic)
	}
}

func TestExposureStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewExposureStore(dir)
	level, ok := s.Get(9999, "tcp")
	if ok {
		t.Error("expected not found")
	}
	if level != ExposureUnknown {
		t.Errorf("expected ExposureUnknown, got %q", level)
	}
}

func TestExposureStore_Remove(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewExposureStore(dir)
	_ = s.Set(443, "tcp", ExposureInternal)
	if err := s.Remove(443, "tcp"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	_, ok := s.Get(443, "tcp")
	if ok {
		t.Error("expected entry to be removed")
	}
}

func TestExposureStore_Remove_NotFound(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewExposureStore(dir)
	if err := s.Remove(1234, "tcp"); err == nil {
		t.Error("expected error when removing non-existent entry")
	}
}

func TestExposureStore_All_SortedByPort(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewExposureStore(dir)
	_ = s.Set(9000, "tcp", ExposurePrivate)
	_ = s.Set(80, "tcp", ExposurePublic)
	_ = s.Set(3000, "udp", ExposureInternal)
	all := s.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(all))
	}
	if all[0].Port != 80 || all[1].Port != 3000 || all[2].Port != 9000 {
		t.Errorf("unexpected order: %v", all)
	}
}

func TestExposureStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	s1, _ := NewExposureStore(dir)
	_ = s1.Set(22, "tcp", ExposurePrivate)

	s2, err := NewExposureStore(dir)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	level, ok := s2.Get(22, "tcp")
	if !ok {
		t.Fatal("expected entry after reload")
	}
	if level != ExposurePrivate {
		t.Errorf("got %q, want %q", level, ExposurePrivate)
	}
}

func TestParseExposureLevel_Valid(t *testing.T) {
	for _, tc := range []struct {
		input    string
		expected ExposureLevel
	}{
		{"public", ExposurePublic},
		{"internal", ExposureInternal},
		{"private", ExposurePrivate},
		{"unknown", ExposureUnknown},
	} {
		got, err := ParseExposureLevel(tc.input)
		if err != nil {
			t.Errorf("ParseExposureLevel(%q): unexpected error: %v", tc.input, err)
		}
		if got != tc.expected {
			t.Errorf("got %q, want %q", got, tc.expected)
		}
	}
}

func TestParseExposureLevel_Invalid(t *testing.T) {
	_, err := ParseExposureLevel("classified")
	if err == nil {
		t.Error("expected error for invalid exposure level")
	}
}
