package portscanner

import (
	"os"
	"testing"
)

func TestVisibilityStore_Set_And_Get(t *testing.T) {
	dir := t.TempDir()
	vs, err := NewVisibilityStore(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := vs.Set(8080, "tcp", VisibilityPublic); err != nil {
		t.Fatalf("Set: %v", err)
	}
	v, ok := vs.Get(8080, "tcp")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if v != VisibilityPublic {
		t.Errorf("got %q, want %q", v, VisibilityPublic)
	}
}

func TestVisibilityStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	vs, _ := NewVisibilityStore(dir)
	_, ok := vs.Get(9999, "tcp")
	if ok {
		t.Error("expected no entry")
	}
}

func TestVisibilityStore_Remove(t *testing.T) {
	dir := t.TempDir()
	vs, _ := NewVisibilityStore(dir)
	_ = vs.Set(443, "tcp", VisibilityInternal)
	if err := vs.Remove(443, "tcp"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	_, ok := vs.Get(443, "tcp")
	if ok {
		t.Error("expected entry to be removed")
	}
}

func TestVisibilityStore_Remove_NotFound(t *testing.T) {
	dir := t.TempDir()
	vs, _ := NewVisibilityStore(dir)
	if err := vs.Remove(1234, "udp"); err == nil {
		t.Error("expected error removing non-existent entry")
	}
}

func TestVisibilityStore_All_SortedByPort(t *testing.T) {
	dir := t.TempDir()
	vs, _ := NewVisibilityStore(dir)
	_ = vs.Set(9000, "tcp", VisibilityPrivate)
	_ = vs.Set(80, "tcp", VisibilityPublic)
	_ = vs.Set(3000, "tcp", VisibilityInternal)
	keys := vs.AllSorted()
	if len(keys) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(keys))
	}
	if keys[0] != "3000/tcp" || keys[1] != "80/tcp" || keys[2] != "9000/tcp" {
		t.Errorf("unexpected order: %v", keys)
	}
}

func TestVisibilityStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	vs, _ := NewVisibilityStore(dir)
	_ = vs.Set(22, "tcp", VisibilityPrivate)

	vs2, err := NewVisibilityStore(dir)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	v, ok := vs2.Get(22, "tcp")
	if !ok {
		t.Fatal("expected entry after reload")
	}
	if v != VisibilityPrivate {
		t.Errorf("got %q, want %q", v, VisibilityPrivate)
	}
}

func TestParseVisibility_Valid(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  Visibility
	}{
		{"public", VisibilityPublic},
		{"internal", VisibilityInternal},
		{"private", VisibilityPrivate},
	} {
		v, err := ParseVisibility(tc.input)
		if err != nil {
			t.Errorf("ParseVisibility(%q): unexpected error: %v", tc.input, err)
		}
		if v != tc.want {
			t.Errorf("got %q, want %q", v, tc.want)
		}
	}
}

func TestParseVisibility_Invalid(t *testing.T) {
	_, err := ParseVisibility("unknown")
	if err == nil {
		t.Error("expected error for unknown visibility")
	}
}

func TestVisibilityStore_MkdirOnCreate(t *testing.T) {
	base := t.TempDir()
	dir := base + "/nested/visibility"
	_, err := NewVisibilityStore(dir)
	if err != nil {
		t.Fatalf("expected dir creation: %v", err)
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("directory was not created")
	}
}
