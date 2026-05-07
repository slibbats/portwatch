package portscanner

import (
	"os"
	"testing"
	"time"
)

func TestAttestationStore_Set_And_Get(t *testing.T) {
	dir := t.TempDir()
	s := NewAttestationStore(dir)

	a := Attestation{
		Port:       8080,
		Proto:      "tcp",
		AttestedBy: "alice",
		Reason:     "approved dev server",
		AttestedAt: time.Now().UTC(),
	}
	if err := s.Set(a); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got, ok, err := s.Get(8080, "tcp")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !ok {
		t.Fatal("expected attestation to exist")
	}
	if got.AttestedBy != "alice" {
		t.Errorf("AttestedBy = %q, want %q", got.AttestedBy, "alice")
	}
	if got.Reason != "approved dev server" {
		t.Errorf("Reason = %q, want %q", got.Reason, "approved dev server")
	}
}

func TestAttestationStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	s := NewAttestationStore(dir)

	_, ok, err := s.Get(9999, "tcp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected not found")
	}
}

func TestAttestationStore_Remove(t *testing.T) {
	dir := t.TempDir()
	s := NewAttestationStore(dir)

	a := Attestation{Port: 443, Proto: "tcp", AttestedBy: "bob", AttestedAt: time.Now()}
	_ = s.Set(a)
	if err := s.Remove(443, "tcp"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	_, ok, _ := s.Get(443, "tcp")
	if ok {
		t.Fatal("expected attestation to be removed")
	}
}

func TestAttestationStore_All_SortedByPort(t *testing.T) {
	dir := t.TempDir()
	s := NewAttestationStore(dir)

	for _, port := range []int{9090, 80, 443} {
		_ = s.Set(Attestation{Port: port, Proto: "tcp", AttestedBy: "x", AttestedAt: time.Now()})
	}

	all, err := s.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 attestations, got %d", len(all))
	}
	if all[0].Port != 80 || all[1].Port != 443 || all[2].Port != 9090 {
		t.Errorf("unexpected order: %v", all)
	}
}

func TestAttestationStore_IsAttested_Expired(t *testing.T) {
	dir := t.TempDir()
	s := NewAttestationStore(dir)

	past := time.Now().Add(-time.Hour)
	a := Attestation{
		Port: 8443, Proto: "tcp",
		AttestedBy: "carol",
		AttestedAt: time.Now().Add(-2 * time.Hour),
		ExpiresAt:  &past,
	}
	_ = s.Set(a)

	ok, err := s.IsAttested(8443, "tcp")
	if err != nil {
		t.Fatalf("IsAttested: %v", err)
	}
	if ok {
		t.Fatal("expected expired attestation to return false")
	}
}

func TestAttestationStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := NewAttestationStore(dir)

	a := Attestation{
		Port: 5432, Proto: "tcp",
		AttestedBy: "dave",
		Reason:     "postgres is expected",
		AttestedAt: time.Now().UTC().Truncate(time.Second),
	}
	_ = s.Set(a)

	s2 := NewAttestationStore(dir)
	got, ok, err := s2.Get(5432, "tcp")
	if err != nil {
		t.Fatalf("Get after reload: %v", err)
	}
	if !ok {
		t.Fatal("expected attestation after reload")
	}
	if got.Reason != a.Reason {
		t.Errorf("Reason = %q, want %q", got.Reason, a.Reason)
	}
}

func TestAttestationStore_Remove_NotFound(t *testing.T) {
	dir := t.TempDir()
	s := NewAttestationStore(dir)

	// Removing non-existent entry should not error
	if err := s.Remove(1234, "udp"); err != nil {
		t.Fatalf("unexpected error removing non-existent: %v", err)
	}
}

func TestAttestationStore_All_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	s := NewAttestationStore(dir)

	// Remove the file to simulate missing store
	_ = os.Remove(s.path())

	all, err := s.All()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 0 {
		t.Fatalf("expected empty, got %d", len(all))
	}
}
