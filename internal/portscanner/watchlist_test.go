package portscanner

import (
	"os"
	"path/filepath"
	"testing"
)

func makeWatchlistListener(port int, proto string) Listener {
	return Listener{Port: port, Protocol: proto, Address: "0.0.0.0"}
}

func TestWatchlist_Add_And_Contains(t *testing.T) {
	w := NewWatchlist()
	w.Add(8080, "tcp", "nginx", "web server")

	if !w.Contains(8080, "tcp") {
		t.Error("expected watchlist to contain port 8080/tcp")
	}
	if w.Contains(9090, "tcp") {
		t.Error("expected watchlist to not contain port 9090/tcp")
	}
	if w.Contains(8080, "udp") {
		t.Error("expected watchlist to not match 8080/udp")
	}
}

func TestWatchlist_FilterUnwatched(t *testing.T) {
	w := NewWatchlist()
	w.Add(22, "tcp", "sshd", "")
	w.Add(80, "tcp", "nginx", "")

	listeners := []Listener{
		makeWatchlistListener(22, "tcp"),
		makeWatchlistListener(80, "tcp"),
		makeWatchlistListener(3000, "tcp"),
	}

	unwatched := w.FilterUnwatched(listeners)
	if len(unwatched) != 1 {
		t.Fatalf("expected 1 unwatched listener, got %d", len(unwatched))
	}
	if unwatched[0].Port != 3000 {
		t.Errorf("expected port 3000, got %d", unwatched[0].Port)
	}
}

func TestWatchlist_FilterUnwatched_EmptyWatchlist(t *testing.T) {
	w := NewWatchlist()
	listeners := []Listener{
		makeWatchlistListener(443, "tcp"),
	}
	out := w.FilterUnwatched(listeners)
	if len(out) != 1 {
		t.Errorf("expected 1 result, got %d", len(out))
	}
}

func TestSaveAndLoadWatchlist_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "watchlist.json")

	w := NewWatchlist()
	w.Add(22, "tcp", "sshd", "secure shell")
	w.Add(53, "udp", "systemd-resolved", "dns")

	if err := SaveWatchlist(path, w); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := LoadWatchlist(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(loaded.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(loaded.Entries))
	}
	if !loaded.Contains(22, "tcp") {
		t.Error("expected loaded watchlist to contain 22/tcp")
	}
}

func TestLoadWatchlist_FileNotFound(t *testing.T) {
	_, err := LoadWatchlist("/nonexistent/watchlist.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestSaveWatchlist_BadPath(t *testing.T) {
	w := NewWatchlist()
	err := SaveWatchlist("/nonexistent/dir/watchlist.json", w)
	if err == nil {
		t.Error("expected error for bad path")
	}
	_ = os.Remove("/nonexistent/dir/watchlist.json")
}
