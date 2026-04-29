package main

import (
	"testing"

	"github.com/user/portwatch/internal/portscanner"
)

func TestAlertRuleStore_SetAndGet_Integration(t *testing.T) {
	dir := t.TempDir()
	store, err := portscanner.NewAlertRuleStore(dir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	store.Set(portscanner.AlertRule{Port: 6379, Protocol: "tcp", Severity: "high", Message: "redis exposed"})
	if err := store.Save(); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	store2, err := portscanner.NewAlertRuleStore(dir)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	got, ok := store2.Get(6379, "tcp")
	if !ok {
		t.Fatal("expected rule after reload")
	}
	if got.Severity != "high" || got.Message != "redis exposed" {
		t.Errorf("unexpected rule: %+v", got)
	}
}

func TestAlertRuleStore_Remove_Integration(t *testing.T) {
	dir := t.TempDir()
	store, _ := portscanner.NewAlertRuleStore(dir)
	store.Set(portscanner.AlertRule{Port: 5432, Protocol: "tcp", Severity: "critical"})
	_ = store.Save()

	store2, _ := portscanner.NewAlertRuleStore(dir)
	if !store2.Remove(5432, "tcp") {
		t.Fatal("expected remove to succeed")
	}
	_ = store2.Save()

	store3, _ := portscanner.NewAlertRuleStore(dir)
	_, ok := store3.Get(5432, "tcp")
	if ok {
		t.Fatal("expected rule to be absent after remove + reload")
	}
}

func TestAlertRuleStore_List_Integration(t *testing.T) {
	dir := t.TempDir()
	store, _ := portscanner.NewAlertRuleStore(dir)
	store.Set(portscanner.AlertRule{Port: 80, Protocol: "tcp", Severity: "low"})
	store.Set(portscanner.AlertRule{Port: 443, Protocol: "tcp", Severity: "low"})
	_ = store.Save()

	store2, _ := portscanner.NewAlertRuleStore(dir)
	all := store2.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(all))
	}
	if all[0].Port != 80 || all[1].Port != 443 {
		t.Errorf("unexpected order: %v", all)
	}
}
