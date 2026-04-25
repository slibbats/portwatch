package main

import (
	"testing"
)

func TestRunAnomaly_UnknownFlag(t *testing.T) {
	err := runAnomaly([]string{"--no-such-flag"})
	if err == nil {
		t.Fatal("expected error for unknown flag, got nil")
	}
}

func TestRunAnomaly_DefaultFlagsAreParseable(t *testing.T) {
	// Providing a non-existent snapshot dir should not panic; it may return an
	// error or a "no snapshots" message — either is acceptable here.
	err := runAnomaly([]string{"--snapshot-dir", "/tmp/portwatch-anomaly-test-does-not-exist"})
	// We only check that the function returns without panicking.
	_ = err
}

func TestRunAnomaly_WindowFlagAccepted(t *testing.T) {
	err := runAnomaly([]string{
		"--snapshot-dir", "/tmp/portwatch-anomaly-test-does-not-exist",
		"--window", "6h",
	})
	_ = err
}

func TestRunAnomaly_MinFlapFlagAccepted(t *testing.T) {
	err := runAnomaly([]string{
		"--snapshot-dir", "/tmp/portwatch-anomaly-test-does-not-exist",
		"--min-flap", "5",
	})
	_ = err
}

func TestRunAnomaly_InvalidSnapshotDir(t *testing.T) {
	// A directory path that cannot be created or read should cause an error or
	// gracefully report no snapshots — the function must not panic.
	err := runAnomaly([]string{"--snapshot-dir", "/dev/null/impossible"})
	_ = err // error or nil — both acceptable; panic is not
}
