package main

import (
	"testing"

	"github.com/user/portwatch/internal/portscanner"
)

func TestDependencyStore_SetAndGet_Integration(t *testing.T) {
	dir := t.TempDir()
	store, err := portscanner.NewDependencyStore(dir)
	if err != nil {
		t.Fatalf("NewDependencyStore: %v", err)
	}
	refs := []portscanner.DependencyRef{
		{Port: 5432, Proto: "tcp", Note: "database"},
	}
	if err := store.Set(8080, "tcp", refs); err != nil {
		t.Fatalf("Set: %v", err)
	}
	d, ok := store.Get(8080, "tcp")
	if !ok {
		t.Fatal("Get: not found")
	}
	if len(d.DependsOn) != 1 || d.DependsOn[0].Port != 5432 {
		t.Errorf("unexpected deps: %+v", d.DependsOn)
	}
}

func TestDependencyStore_Remove_Integration(t *testing.T) {
	dir := t.TempDir()
	store, _ := portscanner.NewDependencyStore(dir)
	_ = store.Set(9090, "tcp", []portscanner.DependencyRef{{Port: 80, Proto: "tcp"}})
	if err := store.Remove(9090, "tcp"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	_, ok := store.Get(9090, "tcp")
	if ok {
		t.Error("expected entry removed")
	}
}

func TestDependencyStore_List_Integration(t *testing.T) {
	dir := t.TempDir()
	store, _ := portscanner.NewDependencyStore(dir)
	_ = store.Set(80, "tcp", nil)
	_ = store.Set(443, "tcp", []portscanner.DependencyRef{{Port: 80, Proto: "tcp"}})
	all := store.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}

func TestDependencyStore_Persistence_Integration(t *testing.T) {
	dir := t.TempDir()
	store, _ := portscanner.NewDependencyStore(dir)
	_ = store.Set(3000, "tcp", []portscanner.DependencyRef{{Port: 5432, Proto: "tcp", Note: "pg"}})

	store2, err := portscanner.NewDependencyStore(dir)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	d, ok := store2.Get(3000, "tcp")
	if !ok {
		t.Fatal("expected persisted entry")
	}
	if d.DependsOn[0].Note != "pg" {
		t.Errorf("note mismatch: %s", d.DependsOn[0].Note)
	}
}
