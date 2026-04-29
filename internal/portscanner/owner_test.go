package portscanner

import (
	"os"
	"testing"
)

func TestOwnerStore_Set_And_Get(t *testing.T) {
	dir := t.TempDir()
	s, err := NewOwnerStore(dir)
	if err != nil {
		t.Fatalf("NewOwnerStore: %v", err)
	}
	e := OwnerEntry{Port: 8080, Proto: "tcp", Owner: "alice", Team: "platform", Email: "alice@example.com"}
	if err := s.Set(e); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, ok := s.Get(8080, "tcp")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if got.Owner != "alice" || got.Team != "platform" {
		t.Errorf("unexpected entry: %+v", got)
	}
}

func TestOwnerStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewOwnerStore(dir)
	_, ok := s.Get(9999, "tcp")
	if ok {
		t.Error("expected not found")
	}
}

func TestOwnerStore_Remove(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewOwnerStore(dir)
	_ = s.Set(OwnerEntry{Port: 443, Proto: "tcp", Owner: "bob"})
	if err := s.Remove(443, "tcp"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	_, ok := s.Get(443, "tcp")
	if ok {
		t.Error("expected entry to be removed")
	}
}

func TestOwnerStore_All_SortedByPort(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewOwnerStore(dir)
	_ = s.Set(OwnerEntry{Port: 9000, Proto: "tcp", Owner: "charlie"})
	_ = s.Set(OwnerEntry{Port: 80, Proto: "tcp", Owner: "dave"})
	_ = s.Set(OwnerEntry{Port: 443, Proto: "tcp", Owner: "eve"})
	all := s.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(all))
	}
	if all[0].Port != 80 || all[1].Port != 443 || all[2].Port != 9000 {
		t.Errorf("entries not sorted: %v", all)
	}
}

func TestOwnerStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	s1, _ := NewOwnerStore(dir)
	_ = s1.Set(OwnerEntry{Port: 8443, Proto: "tcp", Owner: "frank", Email: "frank@corp.io"})

	s2, err := NewOwnerStore(dir)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	got, ok := s2.Get(8443, "tcp")
	if !ok {
		t.Fatal("entry missing after reload")
	}
	if got.Owner != "frank" || got.Email != "frank@corp.io" {
		t.Errorf("unexpected entry after reload: %+v", got)
	}
}

func TestOwnerStore_NonExistentDir(t *testing.T) {
	dir := t.TempDir()
	nonExistent := dir + "/does/not/exist"
	s, err := NewOwnerStore(nonExistent)
	if err != nil {
		t.Fatalf("unexpected error for missing dir: %v", err)
	}
	_ = s.Set(OwnerEntry{Port: 22, Proto: "tcp", Owner: "sysadmin"})
	if _, err := os.Stat(nonExistent); os.IsNotExist(err) {
		t.Error("expected dir to be created")
	}
}
