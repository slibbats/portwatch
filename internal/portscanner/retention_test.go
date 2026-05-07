package portscanner

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeTestSnapshot(t *testing.T, dir, name string, modTime time.Time) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(`{}`), 0644); err != nil {
		t.Fatalf("write test snapshot: %v", err)
	}
	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatalf("set mod time: %v", err)
	}
}

func TestDefaultRetentionPolicy(t *testing.T) {
	p := DefaultRetentionPolicy()
	if p.MaxAgeDays != 30 {
		t.Errorf("expected MaxAgeDays=30, got %d", p.MaxAgeDays)
	}
	if p.MaxCount != 100 {
		t.Errorf("expected MaxCount=100, got %d", p.MaxCount)
	}
}

func TestSaveAndLoadRetentionPolicy_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	p := RetentionPolicy{MaxAgeDays: 7, MaxCount: 50}
	if err := SaveRetentionPolicy(dir, p); err != nil {
		t.Fatalf("save: %v", err)
	}
	loaded, err := LoadRetentionPolicy(dir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.MaxAgeDays != 7 || loaded.MaxCount != 50 {
		t.Errorf("round-trip mismatch: got %+v", loaded)
	}
}

func TestLoadRetentionPolicy_DefaultsWhenMissing(t *testing.T) {
	dir := t.TempDir()
	p, err := LoadRetentionPolicy(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.MaxAgeDays != 30 {
		t.Errorf("expected default MaxAgeDays=30, got %d", p.MaxAgeDays)
	}
}

func TestApplyRetention_PrunesOldFiles(t *testing.T) {
	dir := t.TempDir()
	old := time.Now().AddDate(0, 0, -60)
	recent := time.Now().AddDate(0, 0, -1)
	writeTestSnapshot(t, dir, "old1.json", old)
	writeTestSnapshot(t, dir, "old2.json", old)
	writeTestSnapshot(t, dir, "recent.json", recent)

	result, err := ApplyRetention(dir, RetentionPolicy{MaxAgeDays: 30, MaxCount: 0})
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if result.Pruned != 2 {
		t.Errorf("expected 2 pruned, got %d", result.Pruned)
	}
	if result.Remaining != 1 {
		t.Errorf("expected 1 remaining, got %d", result.Remaining)
	}
}

func TestApplyRetention_PrunesByCount(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	for i := 0; i < 5; i++ {
		writeTestSnapshot(t, dir, filepath.Base(t.TempDir())+".json", now)
	}

	result, err := ApplyRetention(dir, RetentionPolicy{MaxAgeDays: 365, MaxCount: 3})
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if result.Pruned != 2 {
		t.Errorf("expected 2 pruned by count, got %d", result.Pruned)
	}
}

func TestApplyRetention_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	result, err := ApplyRetention(dir, DefaultRetentionPolicy())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Pruned != 0 {
		t.Errorf("expected 0 pruned, got %d", result.Pruned)
	}
}
