package portscanner

import (
	"os"
	"testing"
)

func TestAlertRuleStore_Set_And_Get(t *testing.T) {
	dir := t.TempDir()
	store, err := NewAlertRuleStore(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rule := AlertRule{Port: 8080, Protocol: "tcp", Severity: "high", Message: "unexpected web server"}
	store.Set(rule)
	got, ok := store.Get(8080, "tcp")
	if !ok {
		t.Fatal("expected rule to be found")
	}
	if got.Severity != "high" || got.Message != "unexpected web server" {
		t.Errorf("unexpected rule content: %+v", got)
	}
}

func TestAlertRuleStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewAlertRuleStore(dir)
	_, ok := store.Get(9999, "tcp")
	if ok {
		t.Fatal("expected no rule to be found")
	}
}

func TestAlertRuleStore_Remove(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewAlertRuleStore(dir)
	store.Set(AlertRule{Port: 22, Protocol: "tcp", Severity: "critical"})
	if !store.Remove(22, "tcp") {
		t.Fatal("expected Remove to return true")
	}
	_, ok := store.Get(22, "tcp")
	if ok {
		t.Fatal("expected rule to be gone after remove")
	}
}

func TestAlertRuleStore_Remove_NotFound(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewAlertRuleStore(dir)
	if store.Remove(1234, "udp") {
		t.Fatal("expected Remove to return false for missing rule")
	}
}

func TestAlertRuleStore_All_SortedByPort(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewAlertRuleStore(dir)
	store.Set(AlertRule{Port: 443, Protocol: "tcp"})
	store.Set(AlertRule{Port: 80, Protocol: "tcp"})
	store.Set(AlertRule{Port: 22, Protocol: "tcp"})
	all := store.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(all))
	}
	if all[0].Port != 22 || all[1].Port != 80 || all[2].Port != 443 {
		t.Errorf("rules not sorted by port: %v", all)
	}
}

func TestAlertRuleStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewAlertRuleStore(dir)
	store.Set(AlertRule{Port: 3306, Protocol: "tcp", Severity: "high", Message: "mysql exposed"})
	if err := store.Save(); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	store2, err := NewAlertRuleStore(dir)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	got, ok := store2.Get(3306, "tcp")
	if !ok {
		t.Fatal("expected rule after reload")
	}
	if got.Message != "mysql exposed" {
		t.Errorf("unexpected message: %s", got.Message)
	}
}

func TestNewAlertRuleStore_MissingFile(t *testing.T) {
	dir := t.TempDir()
	_ = os.Remove(dir + "/alert_rules.json")
	_, err := NewAlertRuleStore(dir)
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
}
