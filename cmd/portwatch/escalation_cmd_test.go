package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/portwatch/internal/portscanner"
)

func TestEscalationStore_SetAndGet_Integration(t *testing.T) {
	dir := t.TempDir()
	s, err := portscanner.NewEscalationStore(dir)
	if err != nil {
		t.Fatalf("NewEscalationStore: %v", err)
	}
	if err := s.Set(8080, "tcp", "ops@example.com", "#ops", "high"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	p, ok := s.Get(8080, "tcp")
	if !ok {
		t.Fatal("expected policy")
	}
	if p.Contact != "ops@example.com" {
		t.Errorf("contact mismatch: %s", p.Contact)
	}
}

func TestEscalationStore_Remove_Integration(t *testing.T) {
	dir := t.TempDir()
	s, _ := portscanner.NewEscalationStore(dir)
	_ = s.Set(9090, "tcp", "dev@example.com", "#dev", "medium")
	if err := s.Remove(9090, "tcp"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	_, ok := s.Get(9090, "tcp")
	if ok {
		t.Error("expected policy to be removed")
	}
}

func TestEscalationStore_List_Integration(t *testing.T) {
	dir := t.TempDir()
	s, _ := portscanner.NewEscalationStore(dir)
	_ = s.Set(80, "tcp", "web@example.com", "#web", "low")
	_ = s.Set(443, "tcp", "sec@example.com", "#sec", "critical")
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 policies, got %d", len(all))
	}
}

func TestEscalationStore_Persistence_Integration(t *testing.T) {
	dir := t.TempDir()
	s1, _ := portscanner.NewEscalationStore(dir)
	_ = s1.Set(22, "tcp", "infra@example.com", "#infra", "critical")

	s2, err := portscanner.NewEscalationStore(dir)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	p, ok := s2.Get(22, "tcp")
	if !ok {
		t.Fatal("expected persisted policy")
	}
	if p.MinLevel != "critical" {
		t.Errorf("expected critical, got %s", p.MinLevel)
	}
	_ = filepath.Join(dir, "escalation.json")
	_ = os.Getenv
}
