package portscanner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func makeListener(addr string, port uint16, proto string) Listener {
	return Listener{Addr: addr, Port: port, Protocol: proto}
}

func TestBaseline_Diff_NoNewListeners(t *testing.T) {
	b := &Baseline{
		Listeners: []Listener{
			makeListener("127.0.0.1", 8080, "tcp"),
			makeListener("0.0.0.0", 443, "tcp"),
		},
	}
	current := []Listener{
		makeListener("127.0.0.1", 8080, "tcp"),
		makeListener("0.0.0.0", 443, "tcp"),
	}
	novel := b.Diff(current)
	if len(novel) != 0 {
		t.Errorf("expected 0 novel listeners, got %d", len(novel))
	}
}

func TestBaseline_Diff_DetectsNewListener(t *testing.T) {
	b := &Baseline{
		Listeners: []Listener{
			makeListener("127.0.0.1", 8080, "tcp"),
		},
	}
	current := []Listener{
		makeListener("127.0.0.1", 8080, "tcp"),
		makeListener("0.0.0.0", 9999, "tcp"),
	}
	novel := b.Diff(current)
	if len(novel) != 1 {
		t.Fatalf("expected 1 novel listener, got %d", len(novel))
	}
	if novel[0].Port != 9999 {
		t.Errorf("expected port 9999, got %d", novel[0].Port)
	}
}

func TestBaseline_Diff_EmptyBaseline(t *testing.T) {
	b := &Baseline{}
	current := []Listener{
		makeListener("0.0.0.0", 80, "tcp"),
		makeListener("0.0.0.0", 443, "tcp"),
	}
	novel := b.Diff(current)
	if len(novel) != 2 {
		t.Errorf("expected 2 novel listeners, got %d", len(novel))
	}
}

func TestSaveAndLoadBaseline_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "baseline.json")

	orig := &Baseline{
		CapturedAt: time.Now().UTC().Truncate(time.Second),
		Listeners: []Listener{
			makeListener("127.0.0.1", 22, "tcp"),
			makeListener("0.0.0.0", 80, "tcp"),
		},
	}

	if err := SaveBaseline(path, orig); err != nil {
		t.Fatalf("SaveBaseline error: %v", err)
	}

	loaded, err := LoadBaseline(path)
	if err != nil {
		t.Fatalf("LoadBaseline error: %v", err)
	}

	origJSON, _ := json.Marshal(orig)
	loadedJSON, _ := json.Marshal(loaded)
	if string(origJSON) != string(loadedJSON) {
		t.Errorf("round-trip mismatch:\n got  %s\n want %s", loadedJSON, origJSON)
	}
}

func TestLoadBaseline_MissingFile(t *testing.T) {
	_, err := LoadBaseline("/nonexistent/path/baseline.json")
	if err == nil {
		t.Error("expected error loading missing file, got nil")
	}
}

func TestSaveBaseline_InvalidPath(t *testing.T) {
	b := &Baseline{}
	err := SaveBaseline("/nonexistent/dir/baseline.json", b)
	if err == nil {
		t.Error("expected error saving to invalid path, got nil")
	}
	_ = os.Remove("/nonexistent/dir/baseline.json")
}
