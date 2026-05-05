package main

import (
	"testing"

	"github.com/user/portwatch/internal/portscanner"
)

func TestScheduleStore_SetAndGet_Integration(t *testing.T) {
	dir := t.TempDir()
	store, err := portscanner.NewScheduleStore(dir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	if err := store.Set(8080, "tcp", "0 * * * *", "web"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	e, ok := store.Get(8080, "tcp")
	if !ok {
		t.Fatal("expected entry")
	}
	if e.Cron != "0 * * * *" || e.Label != "web" {
		t.Errorf("unexpected entry: %+v", e)
	}
}

func TestScheduleStore_Remove_Integration(t *testing.T) {
	dir := t.TempDir()
	store, _ := portscanner.NewScheduleStore(dir)
	_ = store.Set(22, "tcp", "@daily", "ssh")
	if err := store.Remove(22, "tcp"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	_, ok := store.Get(22, "tcp")
	if ok {
		t.Error("expected entry to be gone")
	}
}

func TestScheduleStore_List_Integration(t *testing.T) {
	dir := t.TempDir()
	store, _ := portscanner.NewScheduleStore(dir)
	_ = store.Set(80, "tcp", "@hourly", "http")
	_ = store.Set(443, "tcp", "@daily", "https")
	all := store.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	if all[0].Port != 80 || all[1].Port != 443 {
		t.Errorf("unexpected order: %v", all)
	}
}

func TestScheduleStore_Persistence_Integration(t *testing.T) {
	dir := t.TempDir()
	s1, _ := portscanner.NewScheduleStore(dir)
	_ = s1.Set(3306, "tcp", "*/10 * * * *", "mysql")

	s2, err := portscanner.NewScheduleStore(dir)
	if err != nil {
		t.Fatalf("reload error: %v", err)
	}
	e, ok := s2.Get(3306, "tcp")
	if !ok {
		t.Fatal("expected persisted entry")
	}
	if e.Label != "mysql" {
		t.Errorf("label mismatch: %q", e.Label)
	}
}
