package portscanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIncidentStore_Open_And_All(t *testing.T) {
	dir := t.TempDir()
	s := NewIncidentStore(dir)

	inc, err := s.Open(8080, "tcp", IncidentHigh, "unexpected listener")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if inc.Port != 8080 || inc.Proto != "tcp" {
		t.Errorf("unexpected incident fields: %+v", inc)
	}
	if inc.ResolvedAt != nil {
		t.Error("new incident should not be resolved")
	}

	all, err := s.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 incident, got %d", len(all))
	}
	if all[0].Message != "unexpected listener" {
		t.Errorf("message mismatch: %s", all[0].Message)
	}
}

func TestIncidentStore_Resolve(t *testing.T) {
	dir := t.TempDir()
	s := NewIncidentStore(dir)

	inc, err := s.Open(443, "tcp", IncidentCritical, "tls port exposed")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	if err := s.Resolve(inc.ID); err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	all, err := s.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if all[0].ResolvedAt == nil {
		t.Error("expected incident to be resolved")
	}
}

func TestIncidentStore_OpenOnly_FiltersResolved(t *testing.T) {
	dir := t.TempDir()
	s := NewIncidentStore(dir)

	inc1, _ := s.Open(80, "tcp", IncidentLow, "http")
	inc2, _ := s.Open(22, "tcp", IncidentMedium, "ssh")
	_ = s.Resolve(inc1.ID)
	_ = inc2

	open, err := s.OpenOnly()
	if err != nil {
		t.Fatalf("OpenOnly: %v", err)
	}
	if len(open) != 1 {
		t.Fatalf("expected 1 open incident, got %d", len(open))
	}
	if open[0].Port != 22 {
		t.Errorf("expected port 22, got %d", open[0].Port)
	}
}

func TestIncidentStore_All_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	s := NewIncidentStore(dir)

	all, err := s.All()
	if err != nil {
		t.Fatalf("All on empty dir: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("expected empty, got %d", len(all))
	}
}

func TestIncidentStore_All_NonExistentDir(t *testing.T) {
	s := NewIncidentStore(filepath.Join(os.TempDir(), "portwatch_no_such_dir_incident"))
	all, err := s.All()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if all != nil {
		t.Errorf("expected nil slice, got %v", all)
	}
}

func TestIncidentStore_Resolve_NotFound(t *testing.T) {
	dir := t.TempDir()
	s := NewIncidentStore(dir)

	if err := s.Resolve("nonexistent_id"); err == nil {
		t.Error("expected error resolving nonexistent incident")
	}
}
