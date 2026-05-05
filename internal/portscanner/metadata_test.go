package portscanner

import (
	"os"
	"testing"
)

func TestMetadataStore_Set_And_Get(t *testing.T) {
	dir := t.TempDir()
	store, err := NewMetadataStore(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := store.Set(8080, "tcp", "env", "production"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	val, ok := store.Get(8080, "tcp", "env")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if val != "production" {
		t.Errorf("expected 'production', got %q", val)
	}
}

func TestMetadataStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewMetadataStore(dir)
	_, ok := store.Get(9999, "tcp", "missing")
	if ok {
		t.Error("expected not found")
	}
}

func TestMetadataStore_Remove(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewMetadataStore(dir)
	_ = store.Set(443, "tcp", "tier", "critical")
	if err := store.Remove(443, "tcp", "tier"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	_, ok := store.Get(443, "tcp", "tier")
	if ok {
		t.Error("expected entry to be removed")
	}
}

func TestMetadataStore_Remove_NotFound(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewMetadataStore(dir)
	err := store.Remove(1234, "tcp", "nonexistent")
	if err == nil {
		t.Error("expected error for missing entry")
	}
}

func TestMetadataStore_All_SortedByPortAndKey(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewMetadataStore(dir)
	_ = store.Set(9090, "tcp", "zone", "us-east")
	_ = store.Set(80, "tcp", "team", "platform")
	_ = store.Set(80, "tcp", "env", "staging")
	all := store.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(all))
	}
	if all[0].Port != 80 || all[0].Key != "env" {
		t.Errorf("unexpected first entry: %+v", all[0])
	}
	if all[1].Port != 80 || all[1].Key != "team" {
		t.Errorf("unexpected second entry: %+v", all[1])
	}
	if all[2].Port != 9090 {
		t.Errorf("unexpected third entry: %+v", all[2])
	}
}

func TestMetadataStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	s1, _ := NewMetadataStore(dir)
	_ = s1.Set(5432, "tcp", "owner", "db-team")

	s2, err := NewMetadataStore(dir)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	val, ok := s2.Get(5432, "tcp", "owner")
	if !ok || val != "db-team" {
		t.Errorf("expected 'db-team', got %q (ok=%v)", val, ok)
	}
}

func TestNewMetadataStore_MissingFile(t *testing.T) {
	dir := t.TempDir()
	_ = os.Remove(dir + "/metadata.json")
	store, err := NewMetadataStore(dir)
	if err != nil {
		t.Fatalf("expected no error for missing file: %v", err)
	}
	if len(store.All()) != 0 {
		t.Error("expected empty store")
	}
}
