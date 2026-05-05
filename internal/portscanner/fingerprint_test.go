package portscanner

import (
	"os"
	"testing"
)

func makeFingerprintListener(port int, proto, addr, process string) Listener {
	return Listener{Port: port, Protocol: proto, Address: addr, Process: process}
}

func TestNewFingerprint_SortsEntries(t *testing.T) {
	listeners := []Listener{
		makeFingerprintListener(8080, "tcp", "0.0.0.0", "nginx"),
		makeFingerprintListener(22, "tcp", "0.0.0.0", "sshd"),
		makeFingerprintListener(443, "tcp", "0.0.0.0", "nginx"),
	}
	fp := NewFingerprint(listeners)
	if len(fp.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(fp.Entries))
	}
	if fp.Entries[0].Port != 22 {
		t.Errorf("expected first entry port 22, got %d", fp.Entries[0].Port)
	}
	if fp.Entries[2].Port != 8080 {
		t.Errorf("expected last entry port 8080, got %d", fp.Entries[2].Port)
	}
}

func TestSaveAndLoadFingerprint_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	listeners := []Listener{
		makeFingerprintListener(80, "tcp", "0.0.0.0", "apache"),
		makeFingerprintListener(53, "udp", "127.0.0.1", "dnsmasq"),
	}
	fp := NewFingerprint(listeners)
	if err := SaveFingerprint(fp, dir); err != nil {
		t.Fatalf("SaveFingerprint: %v", err)
	}
	loaded, err := LoadFingerprint(dir)
	if err != nil {
		t.Fatalf("LoadFingerprint: %v", err)
	}
	if len(loaded.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(loaded.Entries))
	}
	if loaded.Entries[0].Process != "dnsmasq" {
		t.Errorf("expected dnsmasq, got %s", loaded.Entries[0].Process)
	}
}

func TestLoadFingerprint_FileNotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := LoadFingerprint(dir)
	if err == nil {
		t.Error("expected error for missing fingerprint file")
	}
}

func TestDiffFingerprint_NoChanges(t *testing.T) {
	listeners := []Listener{
		makeFingerprintListener(22, "tcp", "0.0.0.0", "sshd"),
	}
	base := NewFingerprint(listeners)
	curr := NewFingerprint(listeners)
	added, removed := DiffFingerprint(base, curr)
	if len(added) != 0 || len(removed) != 0 {
		t.Errorf("expected no diff, got added=%d removed=%d", len(added), len(removed))
	}
}

func TestDiffFingerprint_DetectsAddedAndRemoved(t *testing.T) {
	base := NewFingerprint([]Listener{
		makeFingerprintListener(22, "tcp", "0.0.0.0", "sshd"),
		makeFingerprintListener(80, "tcp", "0.0.0.0", "nginx"),
	})
	curr := NewFingerprint([]Listener{
		makeFingerprintListener(22, "tcp", "0.0.0.0", "sshd"),
		makeFingerprintListener(9000, "tcp", "0.0.0.0", "unknown"),
	})
	added, removed := DiffFingerprint(base, curr)
	if len(added) != 1 || added[0].Port != 9000 {
		t.Errorf("expected added port 9000, got %v", added)
	}
	if len(removed) != 1 || removed[0].Port != 80 {
		t.Errorf("expected removed port 80, got %v", removed)
	}
}

func TestSaveFingerprint_CreatesDir(t *testing.T) {
	dir := t.TempDir()
	nested := dir + "/sub/dir"
	fp := NewFingerprint([]Listener{})
	if err := SaveFingerprint(fp, nested); err != nil {
		t.Fatalf("expected dir creation, got: %v", err)
	}
	if _, err := os.Stat(nested); err != nil {
		t.Errorf("expected dir to exist: %v", err)
	}
}
