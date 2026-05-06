package main

import (
	"testing"

	"github.com/user/portwatch/internal/portscanner"
)

func TestReviewStore_SetAndGet_Integration(t *testing.T) {
	dir := t.TempDir()
	s, err := portscanner.NewReviewStore(dir)
	if err != nil {
		t.Fatalf("NewReviewStore: %v", err)
	}
	if err := s.Set(8443, "tcp", portscanner.ReviewApproved, "ops", "verified"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	e, ok := s.Get(8443, "tcp")
	if !ok {
		t.Fatal("expected entry")
	}
	if e.Status != portscanner.ReviewApproved {
		t.Errorf("expected approved, got %s", e.Status)
	}
}

func TestReviewStore_Remove_Integration(t *testing.T) {
	dir := t.TempDir()
	s, _ := portscanner.NewReviewStore(dir)
	_ = s.Set(9090, "tcp", portscanner.ReviewPending, "dev", "")
	if err := s.Remove(9090, "tcp"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	_, ok := s.Get(9090, "tcp")
	if ok {
		t.Error("expected entry removed")
	}
}

func TestReviewStore_List_Integration(t *testing.T) {
	dir := t.TempDir()
	s, _ := portscanner.NewReviewStore(dir)
	_ = s.Set(80, "tcp", portscanner.ReviewApproved, "a", "")
	_ = s.Set(443, "tcp", portscanner.ReviewRejected, "b", "")
	all := s.All()
	if len(all) != 2 {
		t.Errorf("expected 2 entries, got %d", len(all))
	}
}

func TestReviewStore_Persistence_Integration(t *testing.T) {
	dir := t.TempDir()
	s1, _ := portscanner.NewReviewStore(dir)
	_ = s1.Set(22, "tcp", portscanner.ReviewRejected, "security", "ssh blocked")

	s2, err := portscanner.NewReviewStore(dir)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	e, ok := s2.Get(22, "tcp")
	if !ok {
		t.Fatal("expected persisted entry after reload")
	}
	if e.Note != "ssh blocked" {
		t.Errorf("unexpected note: %s", e.Note)
	}
}
