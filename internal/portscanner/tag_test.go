package portscanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTagStore_Add_And_Get(t *testing.T) {
	ts := NewTagStore()
	tag := Tag{Port: 8080, Protocol: "tcp", Label: "dev-server", Note: "local dev"}
	ts.Add(tag)

	got, ok := ts.Get(8080, "tcp")
	if !ok {
		t.Fatal("expected tag to be found")
	}
	if got.Label != "dev-server" {
		t.Errorf("expected label %q, got %q", "dev-server", got.Label)
	}
}

func TestTagStore_Add_UpdatesExisting(t *testing.T) {
	ts := NewTagStore()
	ts.Add(Tag{Port: 443, Protocol: "tcp", Label: "https"})
	ts.Add(Tag{Port: 443, Protocol: "tcp", Label: "https-updated"})

	if len(ts.Tags) != 1 {
		t.Errorf("expected 1 tag, got %d", len(ts.Tags))
	}
	got, _ := ts.Get(443, "tcp")
	if got.Label != "https-updated" {
		t.Errorf("expected updated label, got %q", got.Label)
	}
}

func TestTagStore_Get_NotFound(t *testing.T) {
	ts := NewTagStore()
	_, ok := ts.Get(9999, "tcp")
	if ok {
		t.Error("expected not found")
	}
}

func TestTagStore_Remove(t *testing.T) {
	ts := NewTagStore()
	ts.Add(Tag{Port: 22, Protocol: "tcp", Label: "ssh"})

	removed := ts.Remove(22, "tcp")
	if !removed {
		t.Error("expected Remove to return true")
	}
	_, ok := ts.Get(22, "tcp")
	if ok {
		t.Error("expected tag to be gone after remove")
	}
}

func TestTagStore_Remove_NotFound(t *testing.T) {
	ts := NewTagStore()
	if ts.Remove(1234, "udp") {
		t.Error("expected Remove to return false for missing tag")
	}
}

func TestTagStore_Sorted(t *testing.T) {
	ts := NewTagStore()
	ts.Add(Tag{Port: 8080, Protocol: "tcp", Label: "b"})
	ts.Add(Tag{Port: 22, Protocol: "tcp", Label: "a"})
	ts.Add(Tag{Port: 8080, Protocol: "udp", Label: "c"})

	sorted := ts.Sorted()
	if sorted[0].Port != 22 {
		t.Errorf("expected first port 22, got %d", sorted[0].Port)
	}
	if sorted[1].Protocol != "tcp" {
		t.Errorf("expected tcp before udp for port 8080")
	}
}

func TestSaveAndLoadTagStore_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tags.json")

	ts := NewTagStore()
	ts.Add(Tag{Port: 3000, Protocol: "tcp", Label: "app", Note: "test app"})

	if err := SaveTagStore(path, ts); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := LoadTagStore(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(loaded.Tags) != 1 {
		t.Fatalf("expected 1 tag, got %d", len(loaded.Tags))
	}
	if loaded.Tags[0].Label != "app" {
		t.Errorf("expected label %q, got %q", "app", loaded.Tags[0].Label)
	}
}

func TestLoadTagStore_FileNotFound_ReturnsEmpty(t *testing.T) {
	ts, err := LoadTagStore("/nonexistent/path/tags.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ts.Tags) != 0 {
		t.Errorf("expected empty store, got %d tags", len(ts.Tags))
	}
}

func TestLoadTagStore_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	os.WriteFile(path, []byte("not json{"), 0644)

	_, err := LoadTagStore(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
