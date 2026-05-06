package portscanner

import (
	"os"
	"testing"
)

func TestReviewStore_Set_And_Get(t *testing.T) {
	dir := t.TempDir()
	s, err := NewReviewStore(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Set(8080, "tcp", ReviewApproved, "alice", "looks fine"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	e, ok := s.Get(8080, "tcp")
	if !ok {
		t.Fatal("expected entry, got none")
	}
	if e.Status != ReviewApproved || e.Reviewer != "alice" || e.Note != "looks fine" {
		t.Errorf("unexpected entry: %+v", e)
	}
}

func TestReviewStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewReviewStore(dir)
	_, ok := s.Get(9999, "tcp")
	if ok {
		t.Error("expected not found")
	}
}

func TestReviewStore_Remove(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewReviewStore(dir)
	_ = s.Set(443, "tcp", ReviewPending, "bob", "")
	if err := s.Remove(443, "tcp"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	_, ok := s.Get(443, "tcp")
	if ok {
		t.Error("expected entry to be removed")
	}
}

func TestReviewStore_Remove_NotFound(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewReviewStore(dir)
	if err := s.Remove(1234, "udp"); err == nil {
		t.Error("expected error for missing entry")
	}
}

func TestReviewStore_All_SortedByPort(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewReviewStore(dir)
	_ = s.Set(9000, "tcp", ReviewPending, "x", "")
	_ = s.Set(80, "tcp", ReviewApproved, "y", "")
	_ = s.Set(443, "tcp", ReviewRejected, "z", "")
	all := s.All()
	if len(all) != 3 {
		t.Fatalf("expected 3, got %d", len(all))
	}
	if all[0].Port != 80 || all[1].Port != 443 || all[2].Port != 9000 {
		t.Errorf("unexpected order: %v", all)
	}
}

func TestReviewStore_FilterByStatus(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewReviewStore(dir)
	_ = s.Set(80, "tcp", ReviewApproved, "a", "")
	_ = s.Set(443, "tcp", ReviewPending, "b", "")
	_ = s.Set(8080, "tcp", ReviewApproved, "c", "")
	approved := s.FilterByStatus(ReviewApproved)
	if len(approved) != 2 {
		t.Errorf("expected 2 approved, got %d", len(approved))
	}
}

func TestReviewStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	s1, _ := NewReviewStore(dir)
	_ = s1.Set(22, "tcp", ReviewRejected, "sec", "not allowed")

	s2, err := NewReviewStore(dir)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	e, ok := s2.Get(22, "tcp")
	if !ok {
		t.Fatal("expected persisted entry")
	}
	if e.Status != ReviewRejected || e.Reviewer != "sec" {
		t.Errorf("unexpected entry after reload: %+v", e)
	}
	_ = os.RemoveAll(dir)
}
