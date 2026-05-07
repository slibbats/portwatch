package portscanner_test

import (
	"os"
	"testing"

	"github.com/user/portwatch/internal/portscanner"
)

func TestClassificationStore_Set_And_Get(t *testing.T) {
	dir := t.TempDir()
	cs, err := portscanner.NewClassificationStore(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := cs.Set(443, "tcp", portscanner.ClassificationConfidential, "TLS traffic"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	e, ok := cs.Get(443, "tcp")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Level != portscanner.ClassificationConfidential {
		t.Errorf("expected confidential, got %q", e.Level)
	}
	if e.Rationale != "TLS traffic" {
		t.Errorf("expected rationale 'TLS traffic', got %q", e.Rationale)
	}
}

func TestClassificationStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	cs, _ := portscanner.NewClassificationStore(dir)
	_, ok := cs.Get(9999, "tcp")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestClassificationStore_Remove(t *testing.T) {
	dir := t.TempDir()
	cs, _ := portscanner.NewClassificationStore(dir)
	_ = cs.Set(80, "tcp", portscanner.ClassificationPublic, "")
	if err := cs.Remove(80, "tcp"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	_, ok := cs.Get(80, "tcp")
	if ok {
		t.Fatal("expected entry to be removed")
	}
}

func TestClassificationStore_Remove_NotFound(t *testing.T) {
	dir := t.TempDir()
	cs, _ := portscanner.NewClassificationStore(dir)
	if err := cs.Remove(1234, "udp"); err == nil {
		t.Fatal("expected error for missing entry")
	}
}

func TestClassificationStore_All_SortedByPort(t *testing.T) {
	dir := t.TempDir()
	cs, _ := portscanner.NewClassificationStore(dir)
	_ = cs.Set(8080, "tcp", portscanner.ClassificationInternal, "")
	_ = cs.Set(22, "tcp", portscanner.ClassificationRestricted, "")
	_ = cs.Set(443, "tcp", portscanner.ClassificationConfidential, "")
	all := cs.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(all))
	}
	if all[0].Port != 22 || all[1].Port != 443 || all[2].Port != 8080 {
		t.Errorf("entries not sorted by port: %v", all)
	}
}

func TestClassificationStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	cs, _ := portscanner.NewClassificationStore(dir)
	_ = cs.Set(53, "udp", portscanner.ClassificationInternal, "DNS")

	cs2, err := portscanner.NewClassificationStore(dir)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	e, ok := cs2.Get(53, "udp")
	if !ok {
		t.Fatal("expected persisted entry")
	}
	if e.Level != portscanner.ClassificationInternal {
		t.Errorf("expected internal, got %q", e.Level)
	}
}

func TestParseClassificationLevel_Valid(t *testing.T) {
	for _, lvl := range []string{"public", "internal", "confidential", "restricted"} {
		_, err := portscanner.ParseClassificationLevel(lvl)
		if err != nil {
			t.Errorf("expected valid level %q, got error: %v", lvl, err)
		}
	}
}

func TestParseClassificationLevel_Invalid(t *testing.T) {
	_, err := portscanner.ParseClassificationLevel("top-secret")
	if err == nil {
		t.Fatal("expected error for invalid level")
	}
}

func init() {
	_ = os.Stderr
}
