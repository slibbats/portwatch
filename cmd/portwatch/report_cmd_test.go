package main

import (
	"testing"
)

func TestRunReport_UnknownFlag(t *testing.T) {
	err := runReport([]string{"--unknown-flag"})
	if err == nil {
		t.Fatal("expected error for unknown flag, got nil")
	}
}

func TestRunReport_InvalidSnapshotDir(t *testing.T) {
	err := runReport([]string{
		"--snapshot-dir", "/nonexistent/path/portwatch",
		"--latest",
	})
	if err == nil {
		t.Fatal("expected error for missing snapshot dir, got nil")
	}
}

func TestRunReport_DefaultFlagsAreParseable(t *testing.T) {
	fs := []string{
		"--snapshot-dir", t.TempDir(),
		"--format", "table",
		"--sort", "port",
	}
	// No snapshots exist — should return an error about loading, not a parse error
	err := runReport(fs)
	if err == nil {
		t.Fatal("expected error due to empty snapshot dir, got nil")
	}
	// Ensure it's not a flag parse error
	if err.Error() == "report: flag parse error" {
		t.Fatalf("unexpected flag parse error: %v", err)
	}
}

func TestRunReport_FormatFlagAccepted(t *testing.T) {
	for _, fmt := range []string{"table", "compact"} {
		err := runReport([]string{
			"--snapshot-dir", t.TempDir(),
			"--format", fmt,
		})
		// Error expected (no snapshots), but not a flag parse error
		if err != nil && err.Error() == "report: flag parse error" {
			t.Fatalf("format=%q caused flag parse error: %v", fmt, err)
		}
	}
}
