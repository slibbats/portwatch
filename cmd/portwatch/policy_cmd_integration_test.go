package main

import (
	"testing"

	"github.com/iamcalledned/portwatch/internal/portscanner"
)

func TestPolicyStore_SetAndGet_Integration(t *testing.T) {
	dir := t.TempDir()
	store, err := portscanner.NewPolicyStore(dir)
	if err != nil {
		t.Fatalf("NewPolicyStore: %v", err)
	}
	if err := store.Set(8443, "tcp", "allow", "secure api"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	p, ok := store.Get(8443, "tcp")
	if !ok {
		t.Fatal("expected policy")
	}
	if p.Action != "allow" || p.Reason != "secure api" {
		t.Errorf("unexpected policy: %+v", p)
	}
}

func TestPolicyStore_Remove_Integration(t *testing.T) {
	dir := t.TempDir()
	store, _ := portscanner.NewPolicyStore(dir)
	_ = store.Set(9090, "tcp", "deny", "blocked")
	if err := store.Remove(9090, "tcp"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	_, ok := store.Get(9090, "tcp")
	if ok {
		t.Fatal("expected policy to be gone")
	}
}

func TestPolicyStore_List_Integration(t *testing.T) {
	dir := t.TempDir()
	store, _ := portscanner.NewPolicyStore(dir)
	_ = store.Set(80, "tcp", "allow", "http")
	_ = store.Set(443, "tcp", "allow", "https")
	_ = store.Set(23, "tcp", "deny", "telnet")
	all := store.All()
	if len(all) != 3 {
		t.Fatalf("expected 3, got %d", len(all))
	}
	if all[0].Port != 23 {
		t.Errorf("expected sorted by port, first is %d", all[0].Port)
	}
}

func TestPolicyStore_Persistence_Integration(t *testing.T) {
	dir := t.TempDir()
	s1, _ := portscanner.NewPolicyStore(dir)
	_ = s1.Set(5432, "tcp", "allow", "postgres")

	s2, err := portscanner.NewPolicyStore(dir)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	p, ok := s2.Get(5432, "tcp")
	if !ok {
		t.Fatal("expected persisted policy")
	}
	if p.Reason != "postgres" {
		t.Errorf("unexpected reason: %s", p.Reason)
	}
}
