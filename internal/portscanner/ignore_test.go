package portscanner

import (
	"os"
	"path/filepath"
	"testing"
)

func makeIgnoreListener(port int, proto string) Listener {
	return Listener{Port: port, Protocol: proto, Address: "0.0.0.0"}
}

func TestIgnoreStore_Add_And_Contains(t *testing.T) {
	s := NewIgnoreStore("")
	s.Add(8080, "tcp")
	if !s.Contains(8080, "tcp") {
		t.Error("expected 8080/tcp to be ignored")
	}
	if s.Contains(9090, "tcp") {
		t.Error("expected 9090/tcp to not be ignored")
	}
}

func TestIgnoreStore_Remove(t *testing.T) {
	s := NewIgnoreStore("")
	s.Add(443, "tcp")
	s.Remove(443, "tcp")
	if s.Contains(443, "tcp") {
		t.Error("expected 443/tcp to be removed from ignore list")
	}
}

func TestIgnoreStore_Remove_NotFound(t *testing.T) {
	s := NewIgnoreStore("")
	// Should not panic when removing a non-existent entry
	s.Remove(1234, "udp")
}

func TestIgnoreStore_FilterIgnored(t *testing.T) {
	s := NewIgnoreStore("")
	s.Add(22, "tcp")
	s.Add(53, "udp")

	listeners := []Listener{
		makeIgnoreListener(22, "tcp"),
		makeIgnoreListener(80, "tcp"),
		makeIgnoreListener(53, "udp"),
		makeIgnoreListener(443, "tcp"),
	}

	result := s.FilterIgnored(listeners)
	if len(result) != 2 {
		t.Fatalf("expected 2 listeners, got %d", len(result))
	}
	for _, l := range result {
		if s.Contains(l.Port, l.Protocol) {
			t.Errorf("filtered result contains ignored port %d/%s", l.Port, l.Protocol)
		}
	}
}

func TestIgnoreStore_FilterIgnored_EmptyIgnoreList(t *testing.T) {
	s := NewIgnoreStore("")
	listeners := []Listener{
		makeIgnoreListener(80, "tcp"),
		makeIgnoreListener(443, "tcp"),
	}
	result := s.FilterIgnored(listeners)
	if len(result) != 2 {
		t.Errorf("expected all 2 listeners, got %d", len(result))
	}
}

func TestIgnoreStore_All_ReturnsEntries(t *testing.T) {
	s := NewIgnoreStore("")
	s.Add(8080, "tcp")
	s.Add(5353, "udp")
	all := s.All()
	if len(all) != 2 {
		t.Errorf("expected 2 entries, got %d", len(all))
	}
}

func TestIgnoreStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ignore.json")

	s1 := NewIgnoreStore(path)
	s1.Add(22, "tcp")
	s1.Add(53, "udp")
	if err := s1.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	s2 := NewIgnoreStore(path)
	if err := s2.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if !s2.Contains(22, "tcp") {
		t.Error("expected 22/tcp after load")
	}
	if !s2.Contains(53, "udp") {
		t.Error("expected 53/udp after load")
	}
}

func TestIgnoreStore_Load_FileNotFound(t *testing.T) {
	s := NewIgnoreStore("/nonexistent/path/ignore.json")
	if err := s.Load(); err != nil {
		t.Errorf("expected no error for missing file, got %v", err)
	}
}

func TestIgnoreStore_Load_CorruptFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ignore.json")
	if err := os.WriteFile(path, []byte("not json{"), 0644); err != nil {
		t.Fatal(err)
	}
	s := NewIgnoreStore(path)
	if err := s.Load(); err == nil {
		t.Error("expected error for corrupt file")
	}
}
